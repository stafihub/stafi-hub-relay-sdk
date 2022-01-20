package chain_test

import (
	"encoding/hex"
	"testing"

	"github.com/ChainSafe/log15"
	"github.com/stafiprotocol/rtoken-relay-core/config"
	"github.com/stafiprotocol/rtoken-relay-core/core"
	"github.com/stafiprotocol/stafi-hub-relay-sdk/chain"
)

var (
	logger = log15.Root().New("chain", "testChain")
	cfg    = config.RawChainConfig{
		Name:         "testChain",
		Type:         "stafiHub",
		Rsymbol:      "FIS",
		Endpoint:     "http://127.0.0.1:26657",
		KeystorePath: "/Users/tpkeeper/.stafihub",
		Opts: chain.ConfigOption{
			BlockstorePath: "/Users/tpkeeper/.stafihub",
			StartBlock:     0,
			ChainID:        "testId",
			Denom:          "stake",
			GasPrice:       "0.0001stake",
			Account:        "my-account",
		},
	}
)

func TestNewConnection(t *testing.T) {
	_, err := chain.NewConnection(&cfg, logger)
	if err != nil {
		t.Fatal(err)
	}
}

func TestQuerySnapshot(t *testing.T) {
	c, err := chain.NewConnection(&cfg, logger)
	if err != nil {
		t.Fatal(err)
	}
	shotId, err := hex.DecodeString("")
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := c.QuerySnapshot(shotId)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(snapshot)
}
func TestQueryUnbond(t *testing.T) {
	c, err := chain.NewConnection(&cfg, logger)
	if err != nil {
		t.Fatal(err)
	}
	unbond, err := c.QueryPoolUnbond("atom", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(unbond)
}

func TestChainInitialize(t *testing.T) {
	c := chain.NewChain()
	sysErr := make(chan error)
	err := c.Initialize(&cfg, logger, sysErr)
	if err != nil {
		t.Fatal(err)
	}
	router := core.NewRouter(logger)

	c.SetRouter(router)
	c.Start()
}
