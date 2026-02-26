package pg

import (
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/kit/pgdb"
)

const (
	insTable = "ins"

	insFromAddress   = "from_address"
	insValue         = "value"
	insTransactionID = "transaction_id"
	insVin           = "vin"
)

type insQ struct {
	db *pgdb.DB
}

func (i *insQ) New() data.InsQ {
	return NewInsQ(i.db.Clone())
}

func (i *insQ) Insert(in data.In) error {
	query := squirrel.Insert(insTable).SetMap(map[string]interface{}{
		insFromAddress:   in.FromAddress,
		insValue:         in.Value,
		insTransactionID: in.TransactionID,
		insVin:           in.Vin,
	})

	err := i.db.Exec(query)
	if err != nil {
		if strings.Contains(err.Error(), data.DuplicateErrValue) {
			err = data.ErrAlreadyExists
		}

		return errors.Wrap(err, "failed to execute db query")
	}

	return nil
}

func NewInsQ(db *pgdb.DB) data.InsQ {
	return &insQ{
		db: db.Clone(),
	}
}
