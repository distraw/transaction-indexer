package poller

import "github.com/btcsuite/btcd/rpcclient"

type Poller interface {
	Poll() (string, error)
}

type poller struct {
	client *rpcclient.Client
}

func (p *poller) Poll() (string, error) {
	hash, err := p.client.GetBestBlockHash()
	if err != nil {
		return "", err
	}

	return hash.String(), nil
}

func New(c *rpcclient.Client) Poller {
	return &poller{
		client: c,
	}
}
