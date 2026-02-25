package data

type InputsQ interface {
	New() InputsQ
	Insert(in Input) error
}

type Input struct {
	ID int `db:"id" json:"-"`

	FromAddress string  `db:"from_address" json:"received_from"`
	Value       float64 `db:"value" json:"BTCs"`

	TransactionID int `db:"transaction_id" json:"-"`

	Vin int `db:"vin" json:"vin"`
}
