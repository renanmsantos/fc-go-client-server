package main

import (
	"context"
	"database/sql"
	"encoding/json"
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

	file, err := os.Create("../db/database.db")
	if err != nil {
		fmt.Println("ERROR:", err)
	}
	file.Close()

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

	// Two context: For DB: 10ms, For API: 200ms

	// Endpoint available to client: /cotacao on port 8080
	http.HandleFunc("/cotacao", GetDollarQuotationHandler)
	fmt.Println("Running server on port:8080")
	http.ListenAndServe(":8080", nil)

}

// GetDollarQuotationHandler Call API https://economia.awesomeapi.com.br/json/last/USD-BRL and return JSON to client
func GetDollarQuotationHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	fmt.Println("Request started.")
	defer fmt.Println("Request ended.")

	select {
	case <-time.After(200 * time.Millisecond):
		req, err := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer req.Body.Close()

		res, err := io.ReadAll(req.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var quotation Quotation
		err = json.Unmarshal(res, &quotation)
		if err != nil {
			fmt.Println("ERROR:", err)
			return
		}

		SaveQuotationOnDatabase(quotation)

		fmt.Println("Request processed.")
		json.NewEncoder(w).Encode(quotation)
		return
	case <-ctx.Done():
		fmt.Println("Request cancelled")
		http.Error(w, "Request cancelled by client.", http.StatusRequestTimeout)
		return
	}

}

// SaveQuotationOnDatabase Save information on SQLite
func SaveQuotationOnDatabase(quotation Quotation) error {

	dbCtx := context.Background()
	fmt.Println("Save database started.")
	dbCtx, cancel := context.WithTimeout(dbCtx, 10*time.Millisecond)
	defer cancel()
	defer fmt.Println("Save database ended.")

	select {
	case <-dbCtx.Done():
		fmt.Println("Save database cancelled.")
		fmt.Println(dbCtx.Err())
		return dbCtx.Err()
	case <-time.After(10 * time.Millisecond):
		db, err := sql.Open("sqlite3", "file:../db/database.db")
		if err != nil {
			fmt.Println("ERROR:", err)
		}
		defer db.Close()

		stmt, err := db.Prepare("INSERT INTO quotation(bid) values(?)")
		if err != nil {
			fmt.Println("ERROR:", err)
		}

		_, err = stmt.Exec(quotation.Usdbrl.Bid)
		if err != nil {
			fmt.Println("ERROR:", err)
		}
		fmt.Println("Save database sucessfully")
		return nil
	}

}
