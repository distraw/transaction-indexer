package data

type OutputsQ interface {
	New() OutputsQ
	Insert(out Output) error
	Delete(id int) error
	MarkSpent(txid string, vout int, blockHeight int32) error
	MarkUnspentAboveHeight(blockHeight int32) error
	Get(txid string, vout uint32) (*Output, error)
	Exists(txid string, vout uint32) (bool, error)
}

type Output struct {
	ID int `db:"id" json:"-"`

	Txid string `db:"txid" json:"txid"`
	Vout int    `db:"vout" json:"vout"`

	Value              float64 `db:"value" json:"BTCs"`
	SpentInBlockHeight *int32  `db:"spent_in_block_height" json:"spent_in_block"`

	Address       string `db:"address" json:"send_to"`
	TransactionID int    `db:"transaction_id" json:"-"`
}
