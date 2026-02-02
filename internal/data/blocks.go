package data

type BlocksQ interface {
	Insert(block Block) error
	Exists(hash string) (bool, error)
	Get(hash string) (*Block, error)
	GetHighest() (*Block, error)
}

type Block struct {
	ID     int    `structs:"id" db:"id"`
	Hash   string `structs:"hash" db:"hash"`
	Height int32  `structs:"height" db:"height"`
}
