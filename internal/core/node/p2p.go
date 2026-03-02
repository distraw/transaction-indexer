package node

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/peer"
	"github.com/btcsuite/btcd/wire"
	"github.com/pkg/errors"
)

type p2pNode struct {
	context context.Context

	peer *peer.Peer

	netParams chaincfg.Params

	pendingHeadersCh chan []*wire.BlockHeader
	pendingBlockCh   chan *wire.MsgBlock

	blockNotFoundCh chan struct{}

	onInvCallback func(newBlockHash *chainhash.Hash) error

	host    string
	agent   string
	version string
	network string

	errCh chan error

	onInvMutex sync.Mutex

	listenEvents atomic.Bool
}

func (p *p2pNode) OnHeaders(_ *peer.Peer, msg *wire.MsgHeaders) {
	p.pendingHeadersCh <- msg.Headers
}

func (p *p2pNode) OnNotFound(_ *peer.Peer, _ *wire.MsgNotFound) {
	p.blockNotFoundCh <- struct{}{}
}

func (p *p2pNode) OnBlock(_ *peer.Peer, msg *wire.MsgBlock, _ []byte) {
	p.pendingBlockCh <- msg
}

func (p *p2pNode) OnInv(_ *peer.Peer, msg *wire.MsgInv) {
	if !p.listenEvents.Load() {
		return
	}

	for _, inv := range msg.InvList {
		switch inv.Type {
		case wire.InvTypeBlock:
			go func() {
				p.onInvMutex.Lock()
				defer p.onInvMutex.Unlock()

				err := p.onInvCallback(&inv.Hash)

				if err != nil {
					p.errCh <- err
				}
			}()
		}
	}
}

func (p *p2pNode) Subscribe(context context.Context, from *chainhash.Hash, netParams chaincfg.Params, callback func(newBlockHash *chainhash.Hash) error) error {
	p.pendingHeadersCh = make(chan []*wire.BlockHeader, 1)
	p.pendingBlockCh = make(chan *wire.MsgBlock, 1)
	p.blockNotFoundCh = make(chan struct{}, 1)

	p.context = context
	p.netParams = netParams

	p.onInvCallback = callback

	p.errCh = make(chan error, 1)

	cfg := &peer.Config{
		UserAgentName:    p.agent,
		UserAgentVersion: p.version,
		ChainParams:      &p.netParams,

		Listeners: peer.MessageListeners{
			OnHeaders:  p.OnHeaders,
			OnNotFound: p.OnNotFound,
			OnBlock:    p.OnBlock,
			OnInv:      p.OnInv,
		},
	}

	var err error
	p.peer, err = peer.NewOutboundPeer(cfg, p.host)
	if err != nil {
		return errors.Wrap(err, "failed to create new outbound peer")
	}

	connection, err := net.Dial(p.network, p.host)
	if err != nil {
		return errors.Wrap(err, "failed to dial p2p node")
	}

	p.peer.AssociateConnection(connection)
	defer p.peer.Disconnect()

	select {
	case <-p.context.Done():
		return p.context.Err()
	case err := <-p.errCh:
		return errors.Wrap(err, "p2p callback failed")
	}
}

func (p *p2pNode) ListenEvents() {
	p.listenEvents.Store(true)
}

func (p *p2pNode) GetHeaders(locators []*chainhash.Hash, stop chainhash.Hash) ([]*wire.BlockHeader, error) {
	msg := wire.NewMsgGetHeaders()
	msg.BlockLocatorHashes = locators

	msg.HashStop = stop
	msg.ProtocolVersion = wire.ProtocolVersion

	p.peer.QueueMessage(msg, nil)

	headers := []*wire.BlockHeader{}
	if len(locators) == 1 && locators[0].IsEqual(&chainhash.Hash{}) {
		headers = append(headers, &p.netParams.GenesisBlock.Header)
	}

	select {
	case pendingHeaders := <-p.pendingHeadersCh:
		headers = append(headers, pendingHeaders...)
		return headers, nil
	case <-p.context.Done():
		return nil, p.context.Err()
	case <-time.After(10 * time.Second):
		return nil, errors.New("timeout waiting for headers")
	}
}

func (p *p2pNode) GetBlock(hash *chainhash.Hash) (*wire.MsgBlock, error) {
	msg := wire.NewMsgGetData()

	invVect := wire.NewInvVect(wire.InvTypeBlock, hash)
	msg.AddInvVect(invVect)

	doneChan := make(chan struct{}, 1)
	p.peer.QueueMessage(msg, doneChan)
	<-doneChan

	select {
	case block := <-p.pendingBlockCh:
		return block, nil
	case <-p.blockNotFoundCh:
		return nil, errors.New("block was not found on p2p node")
	case <-p.context.Done():
		return nil, p.context.Err()
	case <-time.After(30 * time.Second):
		return nil, errors.New("timeout waiting for block " + hash.String())
	}
}

func NewP2P(host string, agent string, version string, network string) Node {
	return &p2pNode{
		host:    host,
		agent:   agent,
		version: version,
		network: network,
	}
}
