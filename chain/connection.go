package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ChainSafe/log15"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/stafiprotocol/rtoken-relay-core/config"
	"github.com/stafiprotocol/rtoken-relay-core/core"
	hubClient "github.com/stafiprotocol/stafi-hub-relay-sdk/client"
	stafiHubXLedgerTypes "github.com/stafiprotocol/stafihub/x/ledger/types"
	stafiHubXRvoteTypes "github.com/stafiprotocol/stafihub/x/rvote/types"
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

func (c *Connection) SubmitProposal(content stafiHubXRvoteTypes.Content) (string, error) {
	msg, err := stafiHubXRvoteTypes.NewMsgSubmitProposal(c.client.GetFromAddress(), content)
	if err != nil {
		return "", err
	}

	if err := msg.ValidateBasic(); err != nil {
		return "", err
	}
	txBts, err := c.client.ConstructAndSignTx(msg)
	if err != nil {
		return "", err
	}
	return c.client.BroadcastTx(txBts)
}

func (c *Connection) QuerySnapshot(shotId []byte) (*stafiHubXLedgerTypes.QueryGetSnapshotResponse, error) {
	queryClient := stafiHubXLedgerTypes.NewQueryClient(c.client.Ctx())
	params := &stafiHubXLedgerTypes.QueryGetSnapshotRequest{
		ShotId: shotId,
	}

	cc, err := hubClient.Retry(func() (interface{}, error) {
		return queryClient.GetSnapshot(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXLedgerTypes.QueryGetSnapshotResponse), nil
}

func (c *Connection) QueryPoolUnbond(denom, pool string, era uint32) (*stafiHubXLedgerTypes.QueryGetPoolUnbondResponse, error) {
	queryClient := stafiHubXLedgerTypes.NewQueryClient(c.client.Ctx())
	params := &stafiHubXLedgerTypes.QueryGetPoolUnbondRequest{
		Denom: denom,
		Pool:  pool,
		Era:   era,
	}

	cc, err := hubClient.Retry(func() (interface{}, error) {
		return queryClient.GetPoolUnbond(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXLedgerTypes.QueryGetPoolUnbondResponse), nil
}
