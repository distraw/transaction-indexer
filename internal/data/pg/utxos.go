package pg

import (
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/distraw/transaction-indexer/internal/data"
	"gitlab.com/distributed_lab/kit/pgdb"
)

const (
	utxosTable     = "utxos"
	utxosTxid      = "txid"
	utxosVout      = "vout"
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
		utxosTxid:      utxo.Txid,
		utxosVout:      utxo.Vout,
		utxosBlockID:   utxo.BlockID,
		utxosAddressID: utxo.AddressID,
	})

	err := u.db.Exec(query)
	if err != nil {
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

func (u *utxosQ) Get(txid string, vout int) (*data.Utxo, error) {
	query := squirrel.Select("*").From(utxosTable).Where(squirrel.Eq{
		utxosTxid: txid,
		utxosVout: vout,
	})

	var utxo data.Utxo
	err := u.db.Get(&utxo, query)
	if err != nil {
		return nil, err
	}

	return &utxo, nil
}

func (u *utxosQ) Exists(txid string, vout int) (bool, error) {
	query := fmt.Sprintf(
		"SELECT EXISTS (SELECT 1 FROM %s WHERE %s=$1 AND %s=$2)",
		utxosTable,
		utxosTxid,
		utxosVout,
	)

	var ok bool
	err := u.db.RawDB().
		QueryRow(query, txid, vout).
		Scan(&ok)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func NewUtxosQ(db *pgdb.DB) data.UtxosQ {
	return &utxosQ{
		db: db.Clone(),
	}
}
