package data

type BlocksQ interface {
	Insert(block Block) error
	GetHighest() (Block, error)
}

type Block struct {
	Hash   string `db:"hash"`
	Height int32  `db:"height"`
}
