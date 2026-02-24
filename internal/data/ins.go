package data

type InsQ interface {
	New() InsQ
	Insert(in In) error
}

type In struct {
	ID int `db:"id" json:"-"`

	FromAddress string  `db:"from_address" json:"received_from"`
	Value       float64 `db:"value" json:"BTCs"`

	TransactionID int `db:"transaction_id" json:"-"`

	Vin int `db:"vin" json:"vin"`
}
