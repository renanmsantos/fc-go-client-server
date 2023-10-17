package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Quotation struct {
	Usdbrl struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func main() {
	// Make call to server.go to make dollar quotation only field BID.
	dollarQuotation := GetDollarQuotationFromServer()

	// Context: 300ms to call server.
	// Save response on cotacao.txt, on format: Dólar:{valor}
	SaveOnFile(dollarQuotation)
}

func GetDollarQuotationFromServer() Quotation {
	req, err := http.Get("http://localhost:8080/cotacao")
	if err != nil {
		fmt.Println("ERROR:", err)
	}
	defer req.Body.Close()

	res, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Println("ERROR:", err)
	}

	var quotation Quotation
	err = json.Unmarshal(res, &quotation)
	if err != nil {
		fmt.Println("ERROR:", err)
	}

	return quotation
}

func SaveOnFile(dolar Quotation) {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.WriteString("Dólar: " + dolar.Usdbrl.Bid)
	if err != nil {
		panic(err)
	}
	fmt.Println("Saved quotation on file.")
}
