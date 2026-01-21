package config

import (
	"github.com/btcsuite/btcd/rpcclient"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
)

const (
	rpcclientKey = "rpc"
)

func (c *config) RPCClient() *rpcclient.Client {
	return c.rpcclient.Do(func() interface{} {
		var config struct {
			Host         string `fig:"host"`
			User         string `fig:"user"`
			Pass         string `fig:"pass"`
			HTTPPostMode bool   `fig:"http_post_mode"`
			DisableTLS   bool   `fig:"disable_tls"`
		}

		err := figure.
			Out(&config).
			From(kv.MustGetStringMap(c.getter, rpcclientKey)).
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

		return client
	}).(*rpcclient.Client)
}
