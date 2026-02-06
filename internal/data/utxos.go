package data

type UtxosQ interface {
	New() UtxosQ
	Insert(utxo Utxo) error
	Delete(id int) error
	MarkSpent(txid string, vout int, blockHash string) error
	Get(txid string, vout int) (*Utxo, error)
	Exists(txid string, vout uint32) (bool, error)
}

type Utxo struct {
	ID int `db:"id"`

	Txid string `db:"txid"`
	Vout int    `db:"vout"`

	Value float64 `db:"value"`

	AddressID int `db:"address_id"`
	BlockID   int `db:"block_id"`
}
