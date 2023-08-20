//go:build integration
// +build integration

package txparser_test

import (
	"context"
	"net/http"
	"testing"

	"txparser"
)

func Test_Client_CurrentBlockNumber(t *testing.T) {
	ctx := context.Background()
	client := txparser.NewJSONRPCClient(http.DefaultClient, "https://cloudflare-eth.com")

	number, err := client.CurrentBlockNumber(ctx)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if number == 0 {
		t.Error("number should be more than 0")
	}
}

func Test_Client_GetBlockByNumber(t *testing.T) {
	ctx := context.Background()
	client := txparser.NewJSONRPCClient(http.DefaultClient, "https://cloudflare-eth.com")

	block, err := client.GetBlockByNumber(ctx, 17948861)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(block.Transactions) != 118 {
		t.Errorf("transactions slice should have %d item(s), but has %d", 118, len(block.Transactions))
	}
}
