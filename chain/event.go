package chain

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/stafihub/rtoken-relay-core/common/core"
	stafiHubXLedgerTypes "github.com/stafihub/stafihub/x/ledger/types"
)

const maxUint32 = math.MaxUint32

var (
	ErrEventAttributeNumberUnMatch = errors.New("ErrEventAttributeNumberTooFew")
)

func (l *Listener) processBlockEvents(currentBlock int64) error {
	if currentBlock%100 == 0 {
		l.log.Debug("processEvents", "blockNum", currentBlock)
	}

	txs, err := l.conn.client.GetBlockTxs(currentBlock)
	if err != nil {
		return fmt.Errorf("client.GetBlockTxs failed: %s", err)
	}
	for _, tx := range txs {
		for _, log := range tx.Logs {
			for _, event := range log.Events {
				err := l.processStringEvents(event, currentBlock)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (l *Listener) processStringEvents(event types.StringEvent, blockNumber int64) error {
	l.log.Debug("processStringEvents", "event", event)
	m := core.Message{
		Source:      l.symbol,
		Destination: l.caredSymbol,
	}

	oldState := make(map[string]stafiHubXLedgerTypes.PoolBondState)
	var shotId string
	switch {
	case event.Type == stafiHubXLedgerTypes.EventTypeEraPoolUpdated:
		if len(event.Attributes) != 4 {
			return ErrEventAttributeNumberUnMatch
		}

		lastEra, err := strconv.Atoi(event.Attributes[1].Value)
		if err != nil {
			return err
		}
		if int64(lastEra) > int64(maxUint32) {
			return fmt.Errorf("last era overflow %d", lastEra)
		}
		currentEra, err := strconv.Atoi(event.Attributes[2].Value)
		if err != nil {
			return err
		}
		if int64(currentEra) > int64(maxUint32) {
			return fmt.Errorf("current era overflow %d", currentEra)
		}
		e := core.EventEraPoolUpdated{
			Denom:      event.Attributes[0].Value,
			LastEra:    uint32(lastEra),
			CurrentEra: uint32(currentEra),
			ShotId:     event.Attributes[3].Value,
		}
		if l.caredSymbol != core.RSymbol(e.Denom) {
			return nil
		}
		chainEra, err := l.conn.client.QueryChainEra(string(l.caredSymbol))
		if err != nil {
			return err
		}
		if chainEra.GetEra() != e.CurrentEra {
			return nil
		}

		snapshotRes, err := l.conn.client.QuerySnapshot(e.ShotId)
		if err != nil {
			return err
		}
		if snapshotRes.Shot.BondState != stafiHubXLedgerTypes.EraUpdated {
			return nil
		}
		e.Snapshot = snapshotRes.Shot
		m.Reason = core.ReasonEraPoolUpdatedEvent
		m.Content = e
		oldState[event.Type] = stafiHubXLedgerTypes.EraUpdated
		shotId = e.ShotId

	case event.Type == stafiHubXLedgerTypes.EventTypeBondReported:
		if len(event.Attributes) != 2 {
			return ErrEventAttributeNumberUnMatch
		}

		e := core.EventBondReported{
			Denom:  event.Attributes[0].Value,
			ShotId: event.Attributes[1].Value,
		}
		if l.caredSymbol != core.RSymbol(e.Denom) {
			return nil
		}

		snapshotRes, err := l.conn.client.QuerySnapshot(e.ShotId)
		if err != nil {
			return err
		}
		if snapshotRes.Shot.BondState != stafiHubXLedgerTypes.BondReported {
			return nil
		}
		chainEra, err := l.conn.client.QueryChainEra(string(l.caredSymbol))
		if err != nil {
			return err
		}
		if chainEra.GetEra() != snapshotRes.Shot.GetEra() {
			return nil
		}
		e.Snapshot = snapshotRes.Shot
		m.Reason = core.ReasonBondReportedEvent
		m.Content = e
		oldState[event.Type] = stafiHubXLedgerTypes.BondReported
		shotId = e.ShotId

	case event.Type == stafiHubXLedgerTypes.EventTypeActiveReported:
		if len(event.Attributes) != 2 {
			return ErrEventAttributeNumberUnMatch
		}

		e := core.EventActiveReported{
			Denom:  event.Attributes[0].Value,
			ShotId: event.Attributes[1].Value,
		}
		if l.caredSymbol != core.RSymbol(e.Denom) {
			return nil
		}
		snapshotRes, err := l.conn.client.QuerySnapshot(e.ShotId)
		if err != nil {
			return err
		}
		if snapshotRes.Shot.BondState != stafiHubXLedgerTypes.ActiveReported {
			return nil
		}
		chainEra, err := l.conn.client.QueryChainEra(string(l.caredSymbol))
		if err != nil {
			return err
		}
		if chainEra.GetEra() != snapshotRes.Shot.GetEra() {
			return nil
		}
		unbondRes, err := l.conn.client.QueryPoolUnbond(e.Denom, snapshotRes.Shot.Pool, snapshotRes.Shot.Era)
		if err != nil {
			return err
		}
		e.Snapshot = snapshotRes.Shot
		e.PoolUnbond = unbondRes.Unbond
		m.Reason = core.ReasonActiveReportedEvent
		m.Content = e
		oldState[event.Type] = stafiHubXLedgerTypes.ActiveReported
		shotId = e.ShotId

	case event.Type == stafiHubXLedgerTypes.EventTypeRParamsChanged:
		if len(event.Attributes) != 6 {
			return ErrEventAttributeNumberUnMatch
		}
		denom := event.Attributes[0].Value
		if l.caredSymbol != core.RSymbol(denom) {
			return nil
		}
		rparams, err := l.conn.client.QueryRParams(denom)
		if err != nil {
			return err
		}

		eventRParams := core.EventRParamsChanged{
			Denom:            denom,
			GasPrice:         rparams.RParams.GasPrice,
			EraSeconds:       rparams.RParams.EraSeconds,
			LeastBond:        rparams.RParams.LeastBond,
			Offset:           rparams.RParams.Offset,
			TargetValidators: rparams.RParams.Validators,
		}

		m.Reason = core.ReasonRParamsChangedEvent
		m.Content = eventRParams

	default:
		return nil
	}

	l.log.Info("find event", "msg", m, "block number", blockNumber)
	err := l.submitMessage(&m)
	if err != nil {
		return err
	}
	// no need wait
	if m.Reason == core.ReasonRParamsChangedEvent {
		return nil
	}

	// here we wait until snapshot's bondstate change to another
	for {
		snapshotRes, err := l.conn.client.QuerySnapshot(shotId)
		if err != nil {
			l.log.Warn("QuerySnapshot failed", "err", err)
			time.Sleep(BlockRetryInterval)
			continue
		}
		if snapshotRes.GetShot().BondState == oldState[event.Type] {
			time.Sleep(BlockRetryInterval)
			continue
		}
		break
	}
	return nil
}
