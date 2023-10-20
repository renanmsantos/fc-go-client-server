package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"net/http"
	"os"
	"time"
)

type Quotation struct {
	Usdbrl struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func init() {
	os.MkdirAll("../db", 0700)

	if _, err := os.Stat("../db/database.db"); errors.Is(err, os.ErrNotExist) {
		file, err := os.Create("../db/database.db")
		if err != nil {
			fmt.Println("ERROR:", err)
		}
		file.Close()
	}

	db, err := sql.Open("sqlite3", "file:../db/database.db")
	if err != nil {
		fmt.Println("ERROR:", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS `quotation` (`id` INTEGER PRIMARY KEY AUTOINCREMENT, `bid` VARCHAR(64) NULL )")
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	// Endpoint available to client: /cotacao on port 8080
	http.HandleFunc("/cotacao", GetDollarQuotationHandler)
	fmt.Println("Running server on port:8080")
	http.ListenAndServe(":8080", nil)

}

// GetDollarQuotationHandler Call API https://economia.awesomeapi.com.br/json/last/USD-BRL and return JSON to client
func GetDollarQuotationHandler(w http.ResponseWriter, r *http.Request) {

	// Two context: For DB: 10ms, For API: 200ms
	ctxRequest, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	fmt.Println("Request started.")
	defer cancel()
	defer fmt.Println("Request ended.")

	req, err := http.NewRequestWithContext(ctxRequest, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("Request timeout reached.")
			w.WriteHeader(http.StatusRequestTimeout)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	resp, err := io.ReadAll(res.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var quotation Quotation
	err = json.Unmarshal(resp, &quotation)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = SaveQuotationOnDatabase(quotation)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("Save database timeout reached.")
			w.WriteHeader(http.StatusRequestTimeout)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(quotation)

}

// SaveQuotationOnDatabase Save information on SQLite
func SaveQuotationOnDatabase(quotation Quotation) error {

	dbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	fmt.Println("Save database started.")
	defer cancel()
	defer fmt.Println("Save database ended.")

	db, err := sql.Open("sqlite3", "file:../db/database.db")
	if err != nil {
		return err
	}
	defer db.Close()

	stmt, err := db.PrepareContext(dbCtx, "INSERT INTO quotation(bid) values(?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(dbCtx, quotation.Usdbrl.Bid)
	if err != nil {
		return err
	}

	return nil
}
