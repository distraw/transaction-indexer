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
	utxosTable = "utxos"
	utxosTxid  = "txid"
	utxosVout  = "vout"
	utxosValue = "value"

	utxosSpentInBlock = "spent_in_block"

	utxosBlockID   = "block_id"
	utxosAddressID = "address_id"
)

type utxosQ struct {
	db *pgdb.DB
}

func (u *utxosQ) New() data.UtxosQ {
	return NewUtxosQ(u.db.Clone())
}

func (u *utxosQ) Insert(utxo data.Utxo) error {
	query := squirrel.Insert(utxosTable).SetMap(map[string]interface{}{
		utxosTxid:         utxo.Txid,
		utxosVout:         utxo.Vout,
		utxosValue:        utxo.Value,
		utxosSpentInBlock: nil,
		utxosBlockID:      utxo.BlockID,
		utxosAddressID:    utxo.AddressID,
	})

	err := u.db.Exec(query)
	if err != nil {
		if strings.Contains(err.Error(), data.DuplicateErrValue) {
			err = data.ErrAlreadyExists
		}

		return err
	}

	return nil
}

func (u *utxosQ) Delete(id int) error {
	query := squirrel.Delete(utxosTable).Where(squirrel.Eq{
		"id": id,
	})

	err := u.db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func (u *utxosQ) MarkSpent(txid string, vout int, blockHash string) error {
	query := squirrel.Update(utxosTable).SetMap(map[string]interface{}{
		utxosSpentInBlock: blockHash,
	}).Where(squirrel.Eq{
		utxosTxid: txid,
		utxosVout: vout,
	})

	err := u.db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func (u *utxosQ) MarkUnspentByBlock(blockHash string) error {
	query := squirrel.Update(utxosTable).SetMap(map[string]interface{}{
		utxosSpentInBlock: nil,
	}).Where(squirrel.Eq{
		utxosSpentInBlock: blockHash,
	})

	err := u.db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func (u *utxosQ) Get(txid string, vout int) (*data.Utxo, error) {
	query := squirrel.Select("*").From(utxosTable).Where(squirrel.Eq{
		utxosTxid: txid,
		utxosVout: vout,
	})

	var utxo data.Utxo
	err := u.db.Get(&utxo, query)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, data.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &utxo, nil
}

func (u *utxosQ) Exists(txid string, vout uint32) (bool, error) {
	query := fmt.Sprintf(
		"SELECT EXISTS (SELECT 1 FROM %s WHERE %s=$1 AND %s=$2)",
		utxosTable,
		utxosTxid,
		utxosVout,
	)

	var exists bool
	err := u.db.RawDB().
		QueryRow(query, txid, vout).
		Scan(&exists)
	if err != nil {
		if strings.Contains(err.Error(), data.DuplicateErrValue) {
			return false, data.ErrAlreadyExists
		}

		return false, err
	}

	return exists, nil
}

func NewUtxosQ(db *pgdb.DB) data.UtxosQ {
	return &utxosQ{
		db: db.Clone(),
	}
}
