package chain

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/stafihub/rtoken-relay-core/common/config"
	"github.com/stafihub/rtoken-relay-core/common/core"
	"github.com/stafihub/rtoken-relay-core/common/log"
	hubClient "github.com/stafihub/stafi-hub-relay-sdk/client"
)

type Connection struct {
	symbol core.RSymbol
	client *hubClient.Client
	log    log.Logger
}

func NewConnection(cfg *config.RawChainConfig, option *ConfigOption, log log.Logger) (*Connection, error) {
	fmt.Printf("Will open %s wallet from <%s>. \nPlease ", cfg.Name, cfg.KeystorePath)
	key, err := keyring.New(types.KeyringServiceName(), keyring.BackendFile, cfg.KeystorePath, os.Stdin, hubClient.MakeEncodingConfig().Marshaler)
	if err != nil {
		return nil, err
	}
	client, err := hubClient.NewClient(key, option.Account, option.GasPrice, cfg.EndpointList, log)
	if err != nil {
		return nil, fmt.Errorf("hubClient.NewClient err: %s", err)
	}

	c := Connection{
		symbol: core.RSymbol(cfg.Rsymbol),
		client: client,
		log:    log,
	}
	return &c, nil
}

func (c *Connection) BlockStoreUseAddress() string {
	return c.client.GetFromAddress().String()
}
