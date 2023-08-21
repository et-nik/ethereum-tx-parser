package txparser_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"txparser"
)

func Test_Storage_CuncurrentSave(t *testing.T) {
	// Arrange
	ctx := context.Background()
	storage := txparser.NewInmemoryTransactionsStorage()
	wg := sync.WaitGroup{}

	// Act
	_ = storage.SaveTransactions(ctx, "0x123", []txparser.Transaction{
		{
			BlockNumber: "0x1",
			Hash:        "0xabcd0",
			From:        "0x123",
			To:          "0x234",
		},
	})
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			_ = storage.SaveTransactions(ctx, "0x123", []txparser.Transaction{
				{
					BlockNumber: "0x1",
					Hash:        fmt.Sprintf("0xabcd%d", rand.Uint32()),
					From:        "0x123",
					To:          "0x345",
				},
				{
					BlockNumber: "0x1",
					Hash:        fmt.Sprintf("0xabcd%d", rand.Uint32()),
					From:        "0x123",
					To:          "0x456",
				},
			})
			wg.Done()
		}()
	}
	wg.Wait()

	// Assert
	transactions, err := storage.GetTransactionsByAddress(ctx, "0x123")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(transactions) != 201 {
		t.Errorf("transactions slice should have %d item(s), but has %d", 201, len(transactions))
	}
}

func Test_Storage_BigInt(t *testing.T) {
	ctx := context.Background()
	storage := txparser.NewInmemoryTransactionsStorage()
	err := storage.SaveTransactions(
		ctx,
		"0x85d995eba9763907fdf35cd2034144dd9d53ce32cbec21349d4b12823c6860c5",
		[]txparser.Transaction{
			{
				Hash: "0x85d995eba9763907fdf35cd2034144dd9d53ce32cbec21349d4b12823c6860c6",
			},
		},
	)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	transactions, err := storage.GetTransactionsByAddress(ctx, "0x85d995eba9763907fdf35cd2034144dd9d53ce32cbec21349d4b12823c6860c5")

	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(transactions) != 1 {
		t.Error("transactions slice should have 1 item(s), but has", len(transactions))
		t.FailNow()
	}
	if transactions[0].Hash != "0x85d995eba9763907fdf35cd2034144dd9d53ce32cbec21349d4b12823c6860c6" {
		t.Error("hashes should be equal")
	}
}
