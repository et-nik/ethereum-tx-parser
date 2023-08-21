package txparser

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"sync/atomic"
)

type InmemoryBlockStorage struct {
	blockID atomic.Int64
}

func NewInmemoryBlockStorage() *InmemoryBlockStorage {
	return &InmemoryBlockStorage{}
}

func (s *InmemoryBlockStorage) SaveBlockID(_ context.Context, blockID int) error {
	s.blockID.Store(int64(blockID))
	return nil
}

func (s *InmemoryBlockStorage) GetBlockID(_ context.Context) int {
	return int(s.blockID.Load())
}

func (s *InmemoryBlockStorage) WithDBTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// TODO: Implement transactional storage.
	return fn(ctx)
}

type InmemorySubscriptionsStorage struct {
	addresses sync.Map // map[string]struct{}
}

func NewInmemorySubscriptionsStorage() *InmemorySubscriptionsStorage {
	return &InmemorySubscriptionsStorage{}
}

func (s *InmemorySubscriptionsStorage) PutAddress(_ context.Context, address string) error {
	s.addresses.Store(address, struct{}{})
	return nil
}

func (s *InmemorySubscriptionsStorage) IsAddressExists(_ context.Context, address string) bool {
	_, ok := s.addresses.Load(address)
	return ok
}

type InmemoryTransactionsStorage struct {
	// Transactions hashes by address index
	// map[string][]*big.Int
	transactionsByAddress sync.Map

	// Transactions by hash
	// map[*big.Int]Transaction
	transactions sync.Map

	mu sync.Mutex
}

func NewInmemoryTransactionsStorage() *InmemoryTransactionsStorage {
	return &InmemoryTransactionsStorage{}
}

func (s *InmemoryTransactionsStorage) GetTransactionsByAddress(
	_ context.Context,
	address string,
) ([]Transaction, error) {
	v, ok := s.transactionsByAddress.Load(address)
	if !ok {
		return nil, nil
	}

	transactions, ok := v.([]*big.Int)
	if !ok {
		return nil, errors.New("invalid storage data")
	}

	result := make([]Transaction, 0, len(transactions))
	for _, hash := range transactions {
		tx, ok := s.transactions.Load(hash)
		if !ok {
			return nil, nil
		}
		txValue, ok := tx.(Transaction)
		if !ok {
			return nil, errors.New("invalid storage data")
		}

		result = append(result, txValue)
	}

	return result, nil
}

func (s *InmemoryTransactionsStorage) SaveTransactions(
	_ context.Context,
	address string,
	newTransactions []Transaction,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	transactionHashes := make([]*big.Int, 0, len(newTransactions))

	for _, tx := range newTransactions {
		hash, err := convertHexToNum(tx.Hash)
		if err != nil {
			return err
		}
		s.transactions.Store(hash, tx)
		transactionHashes = append(transactionHashes, hash)
	}

	v, ok := s.transactionsByAddress.Load(address)
	if !ok {
		s.transactionsByAddress.Store(address, transactionHashes)
		return nil
	}

	transactions, ok := v.([]*big.Int)
	if !ok {
		return errors.New("invalid storage data")
	}

	transactions = append(transactions, transactionHashes...)
	s.transactionsByAddress.Store(address, transactions)

	return nil
}

func (s *InmemoryTransactionsStorage) DeleteTransactionsByAddress(_ context.Context, address string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.transactionsByAddress.Load(address)
	if !ok {
		return nil
	}

	transactions, ok := v.([]*big.Int)
	if !ok {
		return errors.New("invalid storage data")
	}

	for _, hash := range transactions {
		s.transactions.Delete(hash)
	}

	s.transactionsByAddress.Delete(address)

	return nil
}

func (s *InmemoryTransactionsStorage) WithDBTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// TODO: Implement transactional storage.
	return fn(ctx)
}
