package txparser

import "context"

type Parser interface {
	// last parsed block
	GetCurrentBlock() int

	// add address to observer
	Subscribe(address string) bool

	// list of inbound or outbound transactions for an address
	GetTransactions(address string) []Transaction
}

type DBTXStorage interface {
	WithDBTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type BlockStorage interface {
	DBTXStorage

	SaveBlockID(ctx context.Context, blockID int) error
	GetBlockID(ctx context.Context) int
}

type TransactionStorage interface {
	DBTXStorage

	GetTransactionsByAddress(ctx context.Context, address string) ([]Transaction, error)
	SaveTransactions(ctx context.Context, address string, transaction []Transaction) error
	DeleteTransactionsByAddress(ctx context.Context, address string) error
}

type SubscriptionsStorage interface {
	PutAddress(ctx context.Context, address string) error
	IsAddressExists(ctx context.Context, address string) bool
}

type Client interface {
	CurrentBlockNumber(ctx context.Context) (int, error)
	GetBlockByNumber(ctx context.Context, number int) (*Block, error)
}
