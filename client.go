package txparser

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"math/rand"
	"net/http"
	"strconv"
)

type JSONRPCCall struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
	ID      string `json:"id"`
}

type JSONRPCResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  any    `json:"result"`
	ID      string `json:"ID"`
}

type JSONRPCClient struct {
	httpClient *http.Client
	host       string
}

func NewJSONRPCClient(httpClient *http.Client, host string) *JSONRPCClient {
	return &JSONRPCClient{
		httpClient: httpClient,
		host:       host,
	}
}

func (c *JSONRPCClient) CurrentBlockNumber(ctx context.Context) (int, error) {
	r, err := c.doRequest(ctx, "eth_blockNumber")
	if err != nil {
		return 0, err
	}

	switch v := r.Result.(type) {
	case string:
		bi, err := convertHexToNum(v)
		if err != nil {
			return 0, err
		}
		return int(bi.Int64()), nil
	default:
		return 0, errors.New("invalid response from api")
	}
}

func (c *JSONRPCClient) GetBlockByNumber(ctx context.Context, number int) (*Block, error) {
	hex := convertNumToHex(number)

	r, err := c.doRequest(ctx, "eth_getBlockByNumber", hex, true)
	if err != nil {
		return nil, err
	}

	rawBlock, ok := r.Result.(map[string]any)
	if !ok {
		return nil, errors.New("invalid response from api")
	}
	blockNumber, ok := rawBlock["number"].(string)
	if !ok {
		return nil, errors.New("invalid structure, invalid block number")
	}
	blockHash, ok := rawBlock["hash"].(string)
	if !ok {
		return nil, errors.New("invalid structure, invalid block hash")
	}
	blockTransactions, ok := rawBlock["transactions"].([]any)
	if !ok {
		return nil, errors.New("invalid structure, invalid block transactions")
	}

	block := Block{
		Number:       blockNumber,
		Hash:         blockHash,
		Transactions: make([]Transaction, 0, len(blockTransactions)),
	}

	for _, transaction := range blockTransactions {
		v, ok := transaction.(map[string]any)
		if !ok {
			continue
		}

		tr, err := convertMapRawToTransaction(v)
		if err != nil {
			return nil, err
		}

		block.Transactions = append(block.Transactions, tr)
	}

	return &block, nil
}

func (c *JSONRPCClient) doRequest(ctx context.Context, method string, params ...any) (*JSONRPCResponse, error) {
	id := randomID()

	payload, err := json.Marshal(JSONRPCCall{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		ID:      id,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.host, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	//nolint:bodyclose
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			log.Print(err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid http status code")
	}

	r := JSONRPCResponse{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&r)
	if err != nil {
		return nil, err
	}

	if r.ID != id {
		return nil, errors.New("id does not match")
	}

	return &r, nil
}

func convertHexToNum(s string) (*big.Int, error) {
	if len(s) <= 2 {
		return nil, errors.New("invalid hex length")
	}
	if s[:2] != "0x" {
		return nil, errors.New("invalid hex")
	}

	bi := new(big.Int)
	bi.SetString(s[2:], 16)

	return bi, nil
}

func convertNumToHex(n int) string {
	return fmt.Sprintf("0x%x", n)
}

func convertMapRawToTransaction(v map[string]any) (Transaction, error) {
	blockNumber, ok := v["blockNumber"].(string)
	if !ok {
		return Transaction{}, errors.New("invalid structure, invalid block number")
	}

	blockHash, ok := v["blockHash"].(string)
	if !ok {
		return Transaction{}, errors.New("invalid structure, invalid block hash")
	}

	hash, ok := v["hash"].(string)
	if !ok {
		return Transaction{}, errors.New("invalid structure, invalid hash")
	}

	from, ok := v["from"].(string)
	if !ok {
		return Transaction{}, errors.New("invalid structure, invalid 'from' value")
	}

	to, ok := v["to"].(string)
	if !ok {
		return Transaction{}, errors.New("invalid structure, invalid 'to' value")
	}

	value, ok := v["value"].(string)
	if !ok {
		return Transaction{}, errors.New("invalid structure, invalid 'value' in transaction")
	}

	return Transaction{
		BlockNumber: blockNumber,
		BlockHash:   blockHash,
		Hash:        hash,
		From:        from,
		To:          to,
		Value:       value,
	}, nil
}

func randomID() string {
	// TODO: Implement uuid generator
	//nolint:gosec
	return strconv.Itoa(int(rand.Uint32()))
}
