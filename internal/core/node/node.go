package node

import (
	"context"
	"errors"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

var (
	ErrNoCommonAncestor = errors.New("no common ancestor was found for given locators")
)

type Mode int

const (
	RPC Mode = iota
	P2P
)

type Node interface {
	Subscribe(context context.Context, from *chainhash.Hash, netParams chaincfg.Params, callback func(newBlockHash *chainhash.Hash) error) error
	ListenEvents()

	GetHeaders(locators []*chainhash.Hash, stop chainhash.Hash) ([]*wire.BlockHeader, error)
	GetBlock(hash *chainhash.Hash) (*wire.MsgBlock, error)
}
