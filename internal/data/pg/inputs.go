package pg

import (
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/distraw/transaction-indexer/internal/data"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/kit/pgdb"
)

const (
	inputsTable = "inputs"

	inputsFromAddress   = "from_address"
	inputsValue         = "value"
	inputsTransactionID = "transaction_id"
	inputsVin           = "vin"
)

type inputsQ struct {
	db *pgdb.DB
}

func (i *inputsQ) New() data.InputsQ {
	return NewInsQ(i.db.Clone())
}

func (i *inputsQ) Insert(input data.Input) error {
	query := squirrel.Insert(inputsTable).SetMap(map[string]interface{}{
		inputsFromAddress:   input.FromAddress,
		inputsValue:         input.Value,
		inputsTransactionID: input.TransactionID,
		inputsVin:           input.Vin,
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

func NewInsQ(db *pgdb.DB) data.InputsQ {
	return &inputsQ{
		db: db.Clone(),
	}
}
