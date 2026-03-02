package pg

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/kit/pgdb"
)

const (
	blocksTable = "blocks"

	blocksHash            = "hash"
	blocksPreviousBlockID = "previous_block_id"
)

type blocksQ struct {
	db *pgdb.DB
}

func (b *blocksQ) Insert(block data.Block) (*int, error) {
	query := squirrel.Insert(blocksTable).SetMap(map[string]interface{}{
		blocksHash:            block.Hash,
		blocksPreviousBlockID: block.PreviousBlockID,
	}).Suffix("RETURNING id")

	var id int
	err := b.db.Get(&id, query)
	if err != nil {
		if strings.Contains(err.Error(), data.DuplicateErrValue) {
			return nil, data.ErrAlreadyExists
		}
		return nil, errors.Wrap(err, "failed to execute db query")
	}

	return &id, nil
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
		return false, errors.Wrap(err, "failed to scan results of a db query")
	}

	return ok, nil
}

func (b *blocksQ) Delete(hash string) error {
	query := squirrel.
		Delete(blocksTable).
		Where(squirrel.Eq{
			blocksHash: hash,
		})

	err := b.db.Exec(query)
	if err != nil {
		return errors.Wrap(err, "failed to execute db query")
	}

	return nil
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
	if errors.Is(err, sql.ErrNoRows) {
		return nil, data.ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute db query")
	}

	return &block, nil
}

func (b *blocksQ) GetByID(id int32) (*data.Block, error) {
	query := squirrel.
		Select("*").
		From(blocksTable).
		Where(squirrel.Eq{
			"id": id,
		})

	var block data.Block
	err := b.db.Get(&block, query)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, data.ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get from blocks table")
	}

	return &block, nil
}

func (b *blocksQ) GetByPreviousBlockID(previousBlockID int32) (*data.Block, error) {
	query := squirrel.
		Select("*").
		From(blocksTable).
		Where(squirrel.Eq{
			blocksPreviousBlockID: previousBlockID,
		})

	var block data.Block
	err := b.db.Get(&block, query)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, data.ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to select from blocks table")
	}

	return &block, nil
}

func (b *blocksQ) GetTip() (*data.Block, error) {
	query := squirrel.
		Select("b.*").
		From(blocksTable + " b").
		LeftJoin(blocksTable + " c ON c." + blocksPreviousBlockID + " = b.id").
		Where("c.id IS NULL")

	var block data.Block
	err := b.db.Get(&block, query)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, data.ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to left join in blocks table")
	}

	return &block, nil
}

func NewBlocksQ(db *pgdb.DB) data.BlocksQ {
	return &blocksQ{
		db: db.Clone(),
	}
}
