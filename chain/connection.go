package chain

import (
	"github.com/ChainSafe/log15"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/stafiprotocol/rtoken-relay-core/config"
	"github.com/stafiprotocol/rtoken-relay-core/core"
	hubClient "github.com/stafiprotocol/rtoken-relay-stafihub/client"
)

type Connection struct {
	url              string
	symbol           core.RSymbol
	validatorTargets []types.ValAddress
	currentHeight    int64
	client           *hubClient.Client
	log              log15.Logger
	stop             <-chan int
}

func NewConnection(cfg *config.RawChainConfig, log log15.Logger, stop <-chan struct{}) (*Connection, error) {
	return nil, nil
}
