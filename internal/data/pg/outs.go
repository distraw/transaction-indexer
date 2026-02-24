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
	outsTable = "outs"
	outsTxid  = "txid"
	outsVout  = "vout"
	outsValue = "value"

	outsSpentInBlockHeight = "spent_in_block_height"

	outsTransactionID = "transaction_id"
	outsAddress       = "address"
)

type outsQ struct {
	db *pgdb.DB
}

func (o *outsQ) New() data.OutsQ {
	return NewOutsQ(o.db.Clone())
}

func (o *outsQ) Insert(out data.Out) error {
	query := squirrel.Insert(outsTable).SetMap(map[string]interface{}{
		outsTxid:               out.Txid,
		outsVout:               out.Vout,
		outsValue:              out.Value,
		outsSpentInBlockHeight: nil,
		outsTransactionID:      out.TransactionID,
		outsAddress:            out.Address,
	})

	err := o.db.Exec(query)
	if err != nil {
		if strings.Contains(err.Error(), data.DuplicateErrValue) {
			err = data.ErrAlreadyExists
		}

		return errors.Wrap(err, "failed to execute db query")
	}

	return nil
}

func (o *outsQ) Delete(id int) error {
	query := squirrel.Delete(outsTable).Where(squirrel.Eq{
		"id": id,
	})

	err := o.db.Exec(query)
	if err != nil {
		return errors.Wrap(err, "failed to execute db query")
	}

	return nil
}

func (o *outsQ) MarkSpent(txid string, vout int, blockHeight int32) error {
	query := squirrel.Update(outsTable).SetMap(map[string]interface{}{
		outsSpentInBlockHeight: blockHeight,
	}).Where(squirrel.Eq{
		outsTxid: txid,
		outsVout: vout,
	})

	err := o.db.Exec(query)
	if err != nil {
		return errors.Wrap(err, "failed to execute db query")
	}

	return nil
}

func (o *outsQ) MarkUnspentAboveHeight(blockHeight int32) error {
	query := squirrel.Update(outsTable).SetMap(map[string]interface{}{
		outsSpentInBlockHeight: nil,
	}).Where(squirrel.Gt{
		outsSpentInBlockHeight: blockHeight,
	})

	err := o.db.Exec(query)
	if err != nil {
		return errors.Wrap(err, "failed to execute db query")
	}

	return nil
}

func (o *outsQ) Get(txid string, vout uint32) (*data.Out, error) {
	query := squirrel.Select("*").From(outsTable).Where(squirrel.Eq{
		outsTxid: txid,
		outsVout: vout,
	})

	var out data.Out
	err := o.db.Get(&out, query)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, data.ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute db query")
	}

	return &out, nil
}

func (o *outsQ) Exists(txid string, vout uint32) (bool, error) {
	query := fmt.Sprintf(
		"SELECT EXISTS (SELECT 1 FROM %s WHERE %s=$1 AND %s=$2)",
		outsTable,
		outsTxid,
		outsVout,
	)

	var exists bool
	err := o.db.RawDB().
		QueryRow(query, txid, vout).
		Scan(&exists)
	if err != nil {
		if strings.Contains(err.Error(), data.DuplicateErrValue) {
			return false, data.ErrAlreadyExists
		}

		return false, errors.Wrap(err, "failed to scan results of raw db query")
	}

	return exists, nil
}

func NewOutsQ(db *pgdb.DB) data.OutsQ {
	return &outsQ{
		db: db.Clone(),
	}
}
