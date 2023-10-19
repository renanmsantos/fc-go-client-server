package main

import (
	"context"
	"encoding/json"
	"fmt"
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

func main() {

	// Context: 300ms to call server.
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	// Make call to server.go to make dollar quotation only field BID.
	dollarQuotation, err := GetDollarQuotationFromServer(ctx)
	if err != nil {
		fmt.Println("ERROR on request: ", err)
		return
	}

	// Save response on cotacao.txt, on format: Dólar:{valor}
	err = SaveOnFile(dollarQuotation)
	if err != nil {
		fmt.Println("ERROR on save file: ", err)
		return
	}
}

func GetDollarQuotationFromServer(ctx context.Context) (Quotation, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		return Quotation{}, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return Quotation{}, err
	}
	defer res.Body.Close()

	resp, err := io.ReadAll(res.Body)
	if err != nil {
		return Quotation{}, err
	}

	var quotation Quotation
	err = json.Unmarshal(resp, &quotation)
	if err != nil {
		return Quotation{}, err
	}

	fmt.Print("Returned request server:" + string(resp))
	return quotation, nil
}

func SaveOnFile(dolar Quotation) error {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString("Dólar: " + dolar.Usdbrl.Bid)
	if err != nil {
		return err
	}
	fmt.Println("Saved quotation on file.")
	return nil
}
