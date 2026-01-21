package config

import (
	"github.com/btcsuite/btcd/rpcclient"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/kit/pgdb"
)

type Config interface {
	comfig.Logger
	comfig.Listenerer
	pgdb.Databaser

	RPCClient() *rpcclient.Client
}

type config struct {
	comfig.Logger
	comfig.Listenerer
	pgdb.Databaser

	rpcclient comfig.Once

	getter kv.Getter
}

func New(getter kv.Getter) Config {
	return &config{
		getter:     getter,
		Logger:     comfig.NewLogger(getter, comfig.LoggerOpts{}),
		Listenerer: comfig.NewListenerer(getter),
		Databaser:  pgdb.NewDatabaser(getter),
	}
}
