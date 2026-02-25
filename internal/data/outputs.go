package data

type OutputsQ interface {
	New() OutputsQ
	Insert(out Output) error
	Delete(id int) error
	MarkSpent(txid string, vout uint32, spentInTransactionID int32) error
	Get(txid string, vout uint32) (*Output, error)
	Exists(txid string, vout uint32) (bool, error)
}

type Output struct {
	ID int `db:"id" json:"-"`

	Txid string `db:"txid" json:"txid"`
	Vout int    `db:"vout" json:"vout"`

	Value                float64 `db:"value" json:"BTCs"`
	SpentInTransactionID *int32  `db:"spent_in_transaction_id" json:"spent_in_tx"`

	Address       string `db:"address" json:"send_to"`
	TransactionID int    `db:"transaction_id" json:"-"`
}
