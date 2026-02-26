package pg

import (
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/kit/pgdb"
)

const (
	transactionsTable     = "transactions"
	transactionsTxid      = "txid"
	transactionsBlockID   = "block_id"
	transactionsLocktime  = "locktime"
	transactionsTimestamp = "timestamp"
)

type transactionsQ struct {
	db *pgdb.DB
}

func (t *transactionsQ) New() data.TransactionsQ {
	return NewTransactionsQ(t.db.Clone())
}

func (t *transactionsQ) Insert(transaction data.Transaction) (*int, error) {
	query := squirrel.Insert(transactionsTable).SetMap(map[string]interface{}{
		transactionsTxid:      transaction.Txid,
		transactionsBlockID:   transaction.BlockID,
		transactionsLocktime:  transaction.Locktime,
		transactionsTimestamp: transaction.Timestamp,
	}).Suffix("RETURNING id")

	var id int
	err := t.db.Get(&id, query)
	if err != nil {
		if strings.Contains(err.Error(), data.DuplicateErrValue) {
			return nil, data.ErrAlreadyExists
		}
		return nil, errors.Wrap(err, "failed to execute db query")
	}

	return &id, nil
}

func NewTransactionsQ(db *pgdb.DB) data.TransactionsQ {
	return &transactionsQ{
		db: db.Clone(),
	}
}
