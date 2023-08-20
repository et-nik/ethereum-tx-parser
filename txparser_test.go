package txparser_test

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"testing"
	"time"

	"txparser"
)

// TODO: Improve tests.
// Clarify github.com/stretchr/testify usage ability.
func Test_Parser(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	blockStorage := txparser.NewInmemoryBlockStorage()
	txStorage := txparser.NewInmemoryTransactionsStorage()
	subscriptionsStorage := txparser.NewInmemorySubscriptionsStorage()
	parser := txparser.NewTXParser(
		blockStorage,
		txStorage,
		subscriptionsStorage,
		newClient(),
	)

	go parser.RunWorker(ctx, 1*time.Second)

	parser.Subscribe("0x123")
	time.Sleep(2 * time.Second)

	// Act
	transactions := parser.GetTransactions("0x123")
	if len(transactions) != 1 {
		t.Errorf("transactions slice should have %d item(s), but has %d", 1, len(transactions))
		t.FailNow()
	}

	// Assert
	if !areStructsEqual(transactions[0], txparser.Transaction{
		BlockNumber: "0x2",
		Hash:        "0xabc20",
		From:        "0x123",
		To:          "0x321",
	}) {
		t.Error("structures should be equal")
		t.FailNow()
	}

	// Act #2
	time.Sleep(3 * time.Second)
	transactions = parser.GetTransactions("0x123")
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Hash < transactions[j].Hash
	})

	// Assert
	if len(transactions) != 3 {
		t.Errorf("transactions slice should have %d item(s), but has %d", 3, len(transactions))
		t.FailNow()
	}
	if !areSlicesEqual([]txparser.Transaction{
		{
			BlockNumber: "0x3",
			Hash:        "0xabc30",
			From:        "0x456",
			To:          "0x123",
		},
		{
			BlockNumber: "0x3",
			Hash:        "0xabc32",
			From:        "0x123",
			To:          "0x678",
		},
		{
			BlockNumber: "0x4",
			Hash:        "0xabc40",
			From:        "0x789",
			To:          "0x123",
		},
	}, transactions) {
		t.Error("transactions slices should be equal")
		t.FailNow()
	}
}

type client struct {
	start time.Time
}

func newClient() *client {
	return &client{start: time.Now()}
}

func (c *client) CurrentBlockNumber(_ context.Context) (int, error) {
	if time.Since(c.start) < 500*time.Millisecond {
		return 1, nil
	} else if time.Since(c.start) < 4*time.Second {
		return 2, nil
	}

	return 4, nil
}

func (c *client) GetBlockByNumber(_ context.Context, number int) (*txparser.Block, error) {
	if number == 1 {
		return &txparser.Block{
			Number: "0x1",
			Transactions: []txparser.Transaction{
				{
					BlockNumber: "0x1",
					Hash:        "0xabc10",
					From:        "0x123",
					To:          "0x321",
				},
			},
		}, nil
	}

	if number == 2 {
		return &txparser.Block{
			Number: "0x2",
			Transactions: []txparser.Transaction{
				{
					BlockNumber: "0x2",
					Hash:        "0xabc20",
					From:        "0x123",
					To:          "0x321",
				},
				{
					BlockNumber: "0x2",
					Hash:        "0xabc21",
					From:        "0x321",
					To:          "0x1337",
				},
			},
		}, nil
	}

	if number == 3 {
		return &txparser.Block{
			Number: "0x3",
			Transactions: []txparser.Transaction{
				{
					BlockNumber: "0x3",
					Hash:        "0xabc30",
					From:        "0x456",
					To:          "0x123",
				},
				{
					BlockNumber: "0x3",
					Hash:        "0xabc31",
					From:        "0x321",
					To:          "0x1337",
				},
				{
					BlockNumber: "0x3",
					Hash:        "0xabc32",
					From:        "0x123",
					To:          "0x678",
				},
			},
		}, nil
	}

	if number == 4 {
		return &txparser.Block{
			Number: "0x4",
			Transactions: []txparser.Transaction{
				{
					BlockNumber: "0x4",
					Hash:        "0xabc40",
					From:        "0x789",
					To:          "0x123",
				},
			},
		}, nil
	}

	return nil, errors.New("invalid block")
}

func areStructsEqual(a, b interface{}) bool {
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}

	valueOfA := reflect.ValueOf(a)
	valueOfB := reflect.ValueOf(b)

	if valueOfA.Kind() != reflect.Struct {
		return false
	}

	numFields := valueOfA.NumField()

	for i := 0; i < numFields; i++ {
		fieldA := valueOfA.Field(i)
		fieldB := valueOfB.Field(i)

		if fieldA.Interface() != fieldB.Interface() {
			return false
		}
	}

	return true
}

func areSlicesEqual(a, b interface{}) bool {
	sliceValueA := reflect.ValueOf(a)
	sliceValueB := reflect.ValueOf(b)

	if sliceValueA.Kind() != reflect.Slice || sliceValueB.Kind() != reflect.Slice {
		return false
	}

	if sliceValueA.Len() != sliceValueB.Len() {
		return false
	}

	for i := 0; i < sliceValueA.Len(); i++ {
		elementA := sliceValueA.Index(i)
		elementB := sliceValueB.Index(i)

		if !reflect.DeepEqual(elementA.Interface(), elementB.Interface()) {
			return false
		}
	}

	return true
}
