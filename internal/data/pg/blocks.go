package pg

import (
	"github.com/Masterminds/squirrel"
	"github.com/distraw/transaction-indexer/internal/data"
	"gitlab.com/distributed_lab/kit/pgdb"
)

const (
	blocksTable = "blocks"

	blocksHash   = "hash"
	blocksHeight = "height"
)

type blocksQ struct {
	db *pgdb.DB
}

func (b *blocksQ) Insert(block data.Block) error {
	query := squirrel.Insert(blocksTable).SetMap(map[string]interface{}{
		blocksHash:   block.Hash,
		blocksHeight: block.Height,
	})

	err := b.db.Exec(query)
	return err
}

func (b *blocksQ) GetHighest() (data.Block, error) {
	query := squirrel.
		Select("*").
		From(blocksTable).
		OrderBy(blocksHeight, "DESC").
		Limit(1)

	var block data.Block
	err := b.db.Get(&block, query)
	if err != nil {
		return data.Block{}, err
	}

	return block, nil
}

func NewBlocksQ(db *pgdb.DB) data.BlocksQ {
	return &blocksQ{
		db: db.Clone(),
	}
}
