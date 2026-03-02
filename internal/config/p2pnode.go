package config

import (
	"github.com/distraw/transaction-indexer/internal/core/node"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
)

const (
	p2pNodeKey = "p2pNode"
)

func (c *config) P2PNode() node.Node {
	return c.p2pNode.Do(func() interface{} {
		var config struct {
			Host    string `fig:"host,required"`
			Agent   string `fig:"agent,required"`
			Version string `fig:"version,required"`
			Network string `fig:"network,required"`
		}

		err := figure.
			Out(&config).
			From(kv.MustGetStringMap(c.getter, p2pNodeKey)).
			Please()
		if err != nil {
			panic(err)
		}

		return node.NewP2P(config.Host, config.Agent, config.Version, config.Network)
	}).(node.Node)
}
