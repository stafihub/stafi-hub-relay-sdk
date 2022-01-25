package chain

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ChainSafe/log15"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/stafiprotocol/rtoken-relay-core/config"
	"github.com/stafiprotocol/rtoken-relay-core/core"
	hubClient "github.com/stafiprotocol/stafi-hub-relay-sdk/client"
)

type Connection struct {
	symbol core.RSymbol
	client *hubClient.Client
	log    log15.Logger
}

func NewConnection(cfg *config.RawChainConfig, log log15.Logger) (*Connection, error) {
	bts, err := json.Marshal(cfg.Opts)
	if err != nil {
		return nil, err
	}
	option := ConfigOption{}
	err = json.Unmarshal(bts, &option)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Will open cosmos wallet from <%s>. \nPlease ", cfg.KeystorePath)
	key, err := keyring.New(types.KeyringServiceName(), keyring.BackendFile, cfg.KeystorePath, os.Stdin)
	if err != nil {
		return nil, err
	}
	client, err := hubClient.NewClient(key, option.ChainID, option.Account, option.GasPrice, option.Denom, cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	c := Connection{
		symbol: core.RSymbol(cfg.Rsymbol),
		client: client,
		log:    log,
	}
	return &c, nil
}
