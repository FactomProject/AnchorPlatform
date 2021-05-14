package wallet

import (
	"fmt"

	"github.com/FactomProject/AnchorPlatform/config"
	"github.com/FactomProject/factom"
)

const (
	ChainECCost = 10
)

type Wallet interface {
	GetEC() *factom.ECAddress
	CommitRevealEntry(entry *factom.Entry) (string, error)
	CommitRevealChain(chain *factom.Chain) (string, error)
}

type Context struct {
	ec *factom.ECAddress
}

func NewWallet(conf *config.Config) (Wallet, error) {

	// setup EC pub-priv keypair from Es address
	ECAddress, err := factom.GetECAddress(conf.Factom.EsAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid Es address set in config %s", conf.Factom.EsAddress)
	}

	return &Context{ECAddress}, nil

}

func (c *Context) GetEC() *factom.ECAddress {
	return c.ec
}

func (c *Context) checkBalance(cost int8) bool {

	balance, _ := factom.GetECBalance(c.ec.PubString())
	return balance >= int64(cost)

}

func (c *Context) CommitRevealEntry(entry *factom.Entry) (string, error) {

	// calculate entry cost
	cost, err := factom.EntryCost(entry)
	if err != nil {
		return "", err
	}

	// check if EC balance enought for tx
	if res := c.checkBalance(cost); !res {
		err = fmt.Errorf("not enough Entry Credits to create entry")
		return "", err
	}

	// commit entry
	_, err = factom.CommitEntry(entry, c.GetEC())
	if err != nil {
		return "", err
	}

	// reveal entry
	resp, err := factom.RevealEntry(entry)
	if err != nil {
		return "", err
	}

	return resp, nil

}

func (c *Context) CommitRevealChain(chain *factom.Chain) (string, error) {

	// calculate entry cost
	cost, err := factom.EntryCost(chain.FirstEntry)
	if err != nil {
		return "", err
	}

	// check if EC balance enought for tx
	if res := c.checkBalance(cost + ChainECCost); !res {
		err = fmt.Errorf("not enough Entry Credits to create chain")
		return "", err
	}

	// commit chain
	_, err = factom.CommitChain(chain, c.GetEC())
	if err != nil {
		return "", err
	}

	// reveal chain
	resp, err := factom.RevealChain(chain)
	if err != nil {
		return "", err
	}

	return resp, nil

}
