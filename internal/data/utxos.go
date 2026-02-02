package data

type UtxosQ interface {
	New() UtxosQ
	Insert(utxo Utxo) error
	Delete(id int) error
	Get(txid string, vout int) (*Utxo, error)
	Exists(txid string, vout int) (bool, error)
}

type Utxo struct {
	ID int `structs:"id" db:"id"`

	Txid string `structs:"txid" db:"txid"`
	Vout int    `structs:"vout" db:"vout"`

	AddressID int `structs:"address_id" db:"address_id"`
	BlockID   int `structs:"block_id" db:"block_id"`
}
