package pg

import (
	"fmt"

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

func (b *blocksQ) Exists(hash string) (bool, error) {
	query := fmt.Sprintf(
		"SELECT EXISTS (SELECT 1 FROM %s WHERE %s=$1)",
		blocksTable,
		blocksHash,
	)

	var ok bool
	err := b.db.RawDB().
		QueryRow(query, hash).
		Scan(&ok)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (b *blocksQ) Get(hash string) (*data.Block, error) {
	query := squirrel.
		Select("*").
		From(blocksTable).
		Where(squirrel.Eq{
			blocksHash: hash,
		})

	var block data.Block
	err := b.db.Get(&block, query)
	if err != nil {
		return nil, err
	}

	return &block, nil
}

func (b *blocksQ) GetHighest() (*data.Block, error) {
	query := squirrel.
		Select("*").
		From(blocksTable).
		OrderBy(blocksHeight, "DESC").
		Limit(1)

	var block data.Block
	err := b.db.Get(&block, query)
	if err != nil {
		return nil, err
	}

	return &block, nil
}

func NewBlocksQ(db *pgdb.DB) data.BlocksQ {
	return &blocksQ{
		db: db.Clone(),
	}
}
