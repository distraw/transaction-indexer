package node

import (
	"context"
	"errors"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

var (
	ErrNoCommonAncestor = errors.New("no common ancestor was found for given locators")
)

type Node interface {
	SubscribeToBlocks(context context.Context, callback func(newBlockHash *chainhash.Hash) error) error

	GetHeaders(locators []*chainhash.Hash, stop *chainhash.Hash) ([]*wire.BlockHeader, error)
	GetBlock(hash *chainhash.Hash) (*wire.MsgBlock, error)
}
