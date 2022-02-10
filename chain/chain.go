package chain

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ChainSafe/log15"
	"github.com/stafihub/rtoken-relay-core/common/config"
	"github.com/stafihub/rtoken-relay-core/common/core"
)

var (
	ErrorTerminated            = errors.New("terminated")
	_               core.Chain = &Chain{}
)

type Chain struct {
	rSymbol     core.RSymbol
	name        string
	conn        *Connection
	listener    *Listener // The listener of this chain
	handler     *Handler  // The writer of the chain
	stop        chan<- struct{}
	initialized bool
}

func NewChain() *Chain {
	return &Chain{rSymbol: core.HubRFIS}
}

func (c *Chain) Initialize(cfg *config.RawChainConfig, logger log15.Logger, sysErr chan<- error) error {
	stop := make(chan struct{})
	bts, err := json.Marshal(cfg.Opts)
	if err != nil {
		return err
	}
	option := ConfigOption{}
	err = json.Unmarshal(bts, &option)
	if err != nil {
		return err
	}
	conn, err := NewConnection(cfg, &option, logger)
	if err != nil {
		return err
	}

	bs, err := NewBlockstore(option.BlockstorePath, conn.BlockStoreUseAddress())
	if err != nil {
		return err
	}

	var startBlk uint64
	startBlk, err = StartBlock(bs, uint64(option.StartBlock))
	if err != nil {
		return err
	}

	l := NewListener(cfg.Name, core.RSymbol(cfg.Rsymbol), core.RSymbol(option.CaredSymbol), startBlk, bs, conn, logger, stop, sysErr)
	h := NewHandler(conn, logger, stop, sysErr)

	c.listener = l
	c.handler = h
	c.conn = conn
	c.name = cfg.Name
	c.initialized = true
	c.stop = stop
	return nil
}

func (c *Chain) Start() error {
	if !c.initialized {
		return fmt.Errorf("chain must be initialized with Initialize()")
	}
	err := c.listener.start()
	if err != nil {
		return err
	}
	err = c.handler.start()
	if err != nil {
		return err
	}
	return nil
}

func (c *Chain) SetRouter(r *core.Router) {
	r.Listen(c.RSymbol(), c.handler)

	c.listener.setRouter(r)
	c.handler.setRouter(r)
}

func (c *Chain) RSymbol() core.RSymbol {
	return c.rSymbol
}

func (c *Chain) Name() string {
	return c.name
}

//stop will stop handler and listener
func (c *Chain) Stop() {
	close(c.stop)
}
