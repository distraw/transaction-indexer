package config

import (
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/distraw/transaction-indexer/internal/core/node"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
)

const (
	rpcNodeKey = "rpcNode"
)

func parsePollFrequency(seconds int) time.Duration {
	const day int = 60 * 60 * 24

	if seconds < 1 || seconds > day {
		panic("poll frequency must be between 1 second and 24 hours (86400 seconds)")
	}

	return time.Duration(seconds) * time.Second
}

func (c *config) RPCNode() node.Node {
	return c.rpcNode.Do(func() interface{} {
		var config struct {
			Host                 string `fig:"host,required"`
			User                 string `fig:"user,required"`
			Pass                 string `fig:"pass,required"`
			HTTPPostMode         bool   `fig:"http_post_mode,required"`
			DisableTLS           bool   `fig:"disable_tls,required"`
			PollFrequencySeconds int    `fig:"poll_frequency_seconds,required"`
		}

		err := figure.
			Out(&config).
			From(kv.MustGetStringMap(c.getter, rpcNodeKey)).
			Please()
		if err != nil {
			panic(err)
		}

		connCfg := &rpcclient.ConnConfig{
			Host:         config.Host,
			User:         config.User,
			Pass:         config.Pass,
			HTTPPostMode: config.HTTPPostMode,
			DisableTLS:   config.DisableTLS,
		}

		client, err := rpcclient.New(connCfg, nil)
		if err != nil {
			panic(err)
		}

		return node.NewRPC(
			client,
			parsePollFrequency(config.PollFrequencySeconds),
		)
	}).(node.Node)
}
