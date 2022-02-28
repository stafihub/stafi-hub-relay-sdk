package chain_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/ChainSafe/log15"
	"github.com/stafihub/rtoken-relay-core/common/config"
	"github.com/stafihub/rtoken-relay-core/common/core"
	"github.com/stafihub/stafi-hub-relay-sdk/chain"
)

var (
	logger = log15.Root().New("chain", "testChain")
	option = chain.ConfigOption{
		BlockstorePath: "/Users/tpkeeper/gowork/stafi/rtoken-relay-core/blockstore",
		StartBlock:     0,
		GasPrice:       "0.0001stake",
		Account:        "my-account",
	}
	cfg = config.RawChainConfig{
		Name:         "testChain",
		Rsymbol:      "RFIS",
		Endpoint:     "http://127.0.0.1:26657",
		KeystorePath: "/Users/tpkeeper/.stafihub",
		Opts:         option,
	}
)

func mockStdin() error {
	content := []byte("tpkeeper\n")
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		return err
	}
	if _, err := tmpfile.Write(content); err != nil {
		return err
	}

	if _, err := tmpfile.Seek(0, 0); err != nil {
		return err
	}

	os.Stdin = tmpfile
	return nil
}

func TestNewConnection(t *testing.T) {
	err := mockStdin()
	if err != nil {
		t.Fatal(err)
	}
	_, err = chain.NewConnection(&cfg, &option, logger)
	if err != nil {
		t.Fatal(err)
	}
}

func TestChainInitialize(t *testing.T) {
	err := mockStdin()
	if err != nil {
		t.Fatal(err)
	}
	c := chain.NewChain()
	sysErr := make(chan error)
	err = c.Initialize(&cfg, logger, sysErr)
	if err != nil {
		t.Fatal(err)
	}
	router := core.NewRouter(logger)

	c.SetRouter(router)
	c.Start()
}
