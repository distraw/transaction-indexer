package data

import "time"

type TransactionsQ interface {
	New() TransactionsQ
	Insert(transaction Transaction) (id *int, err error)
}

type Transaction struct {
	ID        int       `db:"id" json:"-"`
	Txid      string    `db:"txid" json:"txid"`
	BlockID   int       `db:"block_id" json:"-"`
	Locktime  uint32    `db:"locktime" json:"locktime"`
	Timestamp time.Time `db:"timestamp" json:"timestamp"`
}
