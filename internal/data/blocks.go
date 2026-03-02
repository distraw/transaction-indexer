package data

type BlocksQ interface {
	Insert(block Block) (id *int, err error)
	Exists(hash string) (bool, error)
	Delete(hash string) error
	Get(hash string) (*Block, error)
	GetByID(id int32) (*Block, error)
	GetByPreviousBlockID(previousBlockID int32) (*Block, error)
	GetTip() (*Block, error)
}

type Block struct {
	ID              int32  `structs:"id" db:"id"`
	Hash            string `structs:"hash" db:"hash"`
	PreviousBlockID *int32 `structs:"previous_block_id" db:"previous_block_id"`
}
