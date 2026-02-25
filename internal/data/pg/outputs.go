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
	outputsTable = "outputs"
	outputsTxid  = "txid"
	outputsVout  = "vout"
	outputsValue = "value"

	outputsSpentInBlockHeight = "spent_in_block_height"

	outputsTransactionID = "transaction_id"
	outputsAddress       = "address"
)

type outputsQ struct {
	db *pgdb.DB
}

func (o *outputsQ) New() data.OutputsQ {
	return NewOutputsQ(o.db.Clone())
}

func (o *outputsQ) Insert(output data.Output) error {
	query := squirrel.Insert(outputsTable).SetMap(map[string]interface{}{
		outputsTxid:               output.Txid,
		outputsVout:               output.Vout,
		outputsValue:              output.Value,
		outputsSpentInBlockHeight: nil,
		outputsTransactionID:      output.TransactionID,
		outputsAddress:            output.Address,
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

func (o *outputsQ) Delete(id int) error {
	query := squirrel.Delete(outputsTable).Where(squirrel.Eq{
		"id": id,
	})

	err := o.db.Exec(query)
	if err != nil {
		return errors.Wrap(err, "failed to execute db query")
	}

	return nil
}

func (o *outputsQ) MarkSpent(txid string, vout int, blockHeight int32) error {
	query := squirrel.Update(outputsTable).SetMap(map[string]interface{}{
		outputsSpentInBlockHeight: blockHeight,
	}).Where(squirrel.Eq{
		outputsTxid: txid,
		outputsVout: vout,
	})

	err := o.db.Exec(query)
	if err != nil {
		return errors.Wrap(err, "failed to execute db query")
	}

	return nil
}

func (o *outputsQ) MarkUnspentAboveHeight(blockHeight int32) error {
	query := squirrel.Update(outputsTable).SetMap(map[string]interface{}{
		outputsSpentInBlockHeight: nil,
	}).Where(squirrel.Gt{
		outputsSpentInBlockHeight: blockHeight,
	})

	err := o.db.Exec(query)
	if err != nil {
		return errors.Wrap(err, "failed to execute db query")
	}

	return nil
}

func (o *outputsQ) Get(txid string, vout uint32) (*data.Output, error) {
	query := squirrel.Select("*").From(outputsTable).Where(squirrel.Eq{
		outputsTxid: txid,
		outputsVout: vout,
	})

	var output data.Output
	err := o.db.Get(&output, query)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, data.ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute db query")
	}

	return &output, nil
}

func (o *outputsQ) Exists(txid string, vout uint32) (bool, error) {
	query := fmt.Sprintf(
		"SELECT EXISTS (SELECT 1 FROM %s WHERE %s=$1 AND %s=$2)",
		outputsTable,
		outputsTxid,
		outputsVout,
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

func NewOutputsQ(db *pgdb.DB) data.OutputsQ {
	return &outputsQ{
		db: db.Clone(),
	}
}
