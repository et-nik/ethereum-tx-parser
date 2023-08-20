package main

import (
	"context"
	"net/http"
	"time"
	"txparser"
)

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

	// Subscribe and listen
	parser.Subscribe("0xb35903e04589e869f240278d0295210353495b57")

	time.Sleep(1 * time.Minute)

	parser.GetTransactions("0xb35903e04589e869f240278d0295210353495b57")
	parser.GetCurrentBlock()
}
