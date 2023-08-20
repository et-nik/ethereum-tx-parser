package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
	"txparser"
)

type SubscribeRequest struct {
	Address string `json:"address"`
}

func main() {
	ctx := context.Background()

	// Dependencies
	blockStorage := txparser.NewInmemoryBlockStorage()
	transactionsStorage := txparser.NewInmemoryTransactionsStorage()
	subscriptionsStorage := txparser.NewInmemorySubscriptionsStorage()
	client := txparser.NewJSONRPCClient(http.DefaultClient, "https://cloudflare-eth.com")
	parser := txparser.NewTXParser(
		blockStorage,
		transactionsStorage,
		subscriptionsStorage,
		client,
	)

	// Run background job
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go parser.RunWorker(ctx, 1*time.Second)

	http.HandleFunc("/currentBlock", func(w http.ResponseWriter, r *http.Request) {
		block := parser.GetCurrentBlock()
		err := json.NewEncoder(w).Encode(block)
		if err != nil {
			log.Print(err)
		}
	})

	http.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		var in SubscribeRequest

		err := json.NewDecoder(r.Body).Decode(&in)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		success := parser.Subscribe(in.Address)
		err = json.NewEncoder(w).Encode(success)
		if err != nil {
			log.Print(err)
		}
	})

	http.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		address := r.URL.Query().Get("address")
		transactions := parser.GetTransactions(address)
		err := json.NewEncoder(w).Encode(transactions)
		if err != nil {
			log.Print(err)
		}
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Print(err)
	}
}
