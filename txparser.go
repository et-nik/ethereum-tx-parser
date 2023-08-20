package txparser

import (
	"context"
	"log"
	"time"
)

type Block struct {
	Number       string        `json:"number"`
	Hash         string        `json:"hash"`
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	BlockNumber string `json:"blockNumber"`
	BlockHash   string `json:"blockHash"`
	Hash        string `json:"hash"`
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
}

type TXParser struct {
	ctx context.Context //TODO: Ask about context in interface

	blocksStorage       BlockStorage
	transactionsStorage TransactionStorage
	subscriptionStorage SubscriptionsStorage

	client Client

	worker *worker
}

func NewTXParser(
	blockStorage BlockStorage,
	transactionStorage TransactionStorage,
	subscriptionStorage SubscriptionsStorage,
	client Client,
) *TXParser {
	txParser := &TXParser{
		ctx:                 context.Background(),
		blocksStorage:       blockStorage,
		transactionsStorage: transactionStorage,
		subscriptionStorage: subscriptionStorage,
		client:              client,
	}

	txParser.worker = newWorker(txParser.parseProcess)

	return txParser
}

func (p *TXParser) SetContext(ctx context.Context) {
	p.ctx = ctx
}

func (p *TXParser) RunWorker(ctx context.Context, period time.Duration) {
	// We don't need to scan all the blocks from the beginning
	currentBlockNumber, err := p.client.CurrentBlockNumber(ctx)
	if err != nil {
		log.Print(err)
		return
	}
	err = p.blocksStorage.SaveBlockID(ctx, currentBlockNumber)
	if err != nil {
		log.Print(err)
		return
	}

	p.worker.Run(ctx, period)
}

func (p *TXParser) GetCurrentBlock() int {
	return p.blocksStorage.GetBlockID(p.ctx)
}

func (p *TXParser) Subscribe(address string) bool {
	err := p.subscriptionStorage.PutAddress(p.ctx, address)
	if err != nil {
		log.Print(err)
		return false
	}

	return true
}

func (p *TXParser) GetTransactions(address string) []Transaction {
	transactions, err := p.transactionsStorage.GetTransactionsByAddress(p.ctx, address)
	if err != nil {
		log.Print(err)
		return nil
	}

	// TODO: Clarify whether transactions need to be deleted
	err = p.transactionsStorage.DeleteTransactionsByAddress(p.ctx, address)
	if err != nil {
		log.Print(err)
	}

	return transactions
}

func (p *TXParser) parseProcess(ctx context.Context) error {
	currentBlockNumber, err := p.client.CurrentBlockNumber(ctx)
	if err != nil {
		return err
	}

	lastSavedBlockNumber := p.blocksStorage.GetBlockID(ctx)
	if err != nil {
		return err
	}

	if currentBlockNumber <= lastSavedBlockNumber {
		return nil
	}

	for blockID := lastSavedBlockNumber + 1; blockID <= currentBlockNumber; blockID++ {
		err := p.singleBlockProcess(ctx, blockID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *TXParser) singleBlockProcess(ctx context.Context, blockID int) error {
	block, err := p.client.GetBlockByNumber(ctx, blockID)
	if err != nil {
		return err
	}

	return p.transactionsStorage.WithDBTransaction(ctx, func(ctx context.Context) error {
		var err error
		for _, transaction := range block.Transactions {
			if p.subscriptionStorage.IsAddressExists(ctx, transaction.From) {
				err = p.transactionsStorage.SaveTransactions(ctx, transaction.From, []Transaction{transaction})
				if err != nil {
					return err
				}
			}

			if p.subscriptionStorage.IsAddressExists(ctx, transaction.To) {
				err = p.transactionsStorage.SaveTransactions(ctx, transaction.To, []Transaction{transaction})
				if err != nil {
					return err
				}
			}
		}

		err = p.blocksStorage.SaveBlockID(ctx, blockID)
		if err != nil {
			return err
		}

		return nil
	})
}
