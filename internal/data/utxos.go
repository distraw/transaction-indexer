package data

type UtxosQ interface {
	New() UtxosQ
	Insert(utxo Utxo) error
	Delete(id int) error
	MarkSpent(txid string, vout int, blockHeight int32) error
	MarkUnspentAboveHeight(blockHeight int32) error
	Get(txid string, vout int) (*Utxo, error)
	Exists(txid string, vout uint32) (bool, error)
}

type Utxo struct {
	ID int `db:"id" json:"-"`

	Txid string `db:"txid" json:"txid"`
	Vout int    `db:"vout" json:"vout"`

	Value              float64 `db:"value" json:"value"`
	SpentInBlockHeight *int32  `db:"spent_in_block_height" json:"spent_in_block_height"`

	AddressID int `db:"address_id" json:"-"`
	BlockID   int `db:"block_id" json:"-"`
}
