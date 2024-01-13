package chain

import (
	"fmt"
	"math/big"
	"time"

	"github.com/stafihub/rtoken-relay-core/common/core"
	"github.com/stafihub/rtoken-relay-core/common/log"
	"github.com/stafihub/rtoken-relay-core/common/utils"
)

var (
	BlockRetryInterval = time.Second * 6
	BlockRetryLimit    = 50
	BlockConfirmNumber = int64(3)
)

type Listener struct {
	name        string
	symbol      core.RSymbol
	caredSymbol core.RSymbol
	startBlock  uint64
	blockstore  utils.Blockstorer
	conn        *Connection
	router      *core.Router
	log         log.Logger
	stopChan    <-chan struct{}
	sysErrChan  chan<- error
}

func NewListener(name string, symbol, caredSymbol core.RSymbol, startBlock uint64, bs utils.Blockstorer, conn *Connection, log log.Logger, stopChan <-chan struct{}, sysErr chan<- error) *Listener {
	return &Listener{
		name:        name,
		symbol:      symbol,
		caredSymbol: caredSymbol,
		startBlock:  startBlock,
		blockstore:  bs,
		conn:        conn,
		log:         log,
		stopChan:    stopChan,
		sysErrChan:  sysErr,
	}
}

func (l *Listener) setRouter(r *core.Router) {
	l.router = r
}

func (l *Listener) start() error {
	if l.router == nil {
		return fmt.Errorf("must set router with setRouter()")
	}
	latestBlk, err := l.conn.client.GetCurrentBlockHeight()
	if err != nil {
		return err
	}

	if latestBlk < int64(l.startBlock) {
		return fmt.Errorf("starting block (%d) is greater than latest known block (%d)", l.startBlock, latestBlk)
	}
	go func() {
		err := l.pollBlocks()
		if err != nil {
			l.log.Error("Polling blocks failed", "err", err)
			l.sysErrChan <- err
		}
	}()

	return nil
}

func (l *Listener) pollBlocks() error {
	var willDealBlock = l.startBlock
	if willDealBlock == 0 {
		willDealBlock = 1
	}
	var retry = BlockRetryLimit
	for {
		select {
		case <-l.stopChan:
			l.log.Info("pollBlocks receive stop chan, will stop")
			return nil
		default:
			if retry <= 0 {
				return fmt.Errorf("pollBlocks reach retry limit ,symbol: %s", l.symbol)
			}

			getBlockHeightStart := time.Now().Second()
			l.log.Debug("GetCurrentBlockHeight", "start", time.Now().Second())
			latestBlk, err := l.conn.client.GetCurrentBlockHeight()
			if err != nil {
				l.log.Error("Failed to fetch latest blockNumber", "err", err)
				retry--
				time.Sleep(BlockRetryInterval)
				continue
			}
			getBlockHeightEnd := time.Now().Second()
			l.log.Debug("GetCurrentBlockHeight", "use", getBlockHeightEnd-getBlockHeightStart, "willDealBlock", willDealBlock, "latestBlk", latestBlk)

			// Sleep if the block we want comes after the most recently finalized block
			if int64(willDealBlock)+BlockConfirmNumber > latestBlk {
				if willDealBlock%100 == 0 {
					l.log.Trace("Block not yet finalized", "target", willDealBlock, "finalBlk", latestBlk)
				}
				time.Sleep(BlockRetryInterval)
				continue
			}
			processBlockStart := time.Now().Second()
			err = l.processBlockEvents(int64(willDealBlock))
			if err != nil {
				l.log.Error("Failed to process events in block", "block", willDealBlock, "err", err)
				retry--
				time.Sleep(BlockRetryInterval)
				continue
			}
			processBlockEnd := time.Now().Second()
			l.log.Debug("processBlockEvents", "use", processBlockEnd-processBlockStart, "dealBlock", willDealBlock)

			storeBlockStart := time.Now().Second()
			// Write to blockstore
			err = l.blockstore.StoreBlock(new(big.Int).SetUint64(willDealBlock))
			if err != nil {
				l.log.Error("Failed to write to blockstore", "err", err)
			}
			willDealBlock++
			storeBlockEnd := time.Now().Second()
			l.log.Debug("StoreBlock", "use", storeBlockEnd-storeBlockStart, "dealBlock", willDealBlock)

			retry = BlockRetryLimit
		}
	}
}

func (l *Listener) submitMessage(m *core.Message) error {
	if len(m.Source) == 0 || len(m.Destination) == 0 {
		return fmt.Errorf("submitMessage failed, no source or destination %s", m)
	}
	err := l.router.Send(m)
	if err != nil {
		l.log.Error("failed to send message", "err", err, "msg", m)
	}
	return err
}
