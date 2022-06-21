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
	stafiHubXRValidatorTypes "github.com/stafihub/stafihub/x/rvalidator/types"
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
		// return if already dealed
		if chainEra.GetEra() != e.CurrentEra {
			return nil
		}

		snapshotRes, err := l.conn.client.QuerySnapshot(e.ShotId)
		if err != nil {
			return err
		}
		// return if already dealed
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
		// return if already dealed
		if snapshotRes.Shot.BondState != stafiHubXLedgerTypes.BondReported {
			return nil
		}
		chainEra, err := l.conn.client.QueryChainEra(string(l.caredSymbol))
		if err != nil {
			return err
		}
		// return if already dealed
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
		// return if already dealed
		if snapshotRes.Shot.BondState != stafiHubXLedgerTypes.ActiveReported {
			return nil
		}
		chainEra, err := l.conn.client.QueryChainEra(string(l.caredSymbol))
		if err != nil {
			return err
		}
		// return if already dealed
		if chainEra.GetEra() != snapshotRes.Shot.GetEra() {
			return nil
		}
		unbondRes, err := l.conn.client.QueryPoolUnbond(e.Denom, snapshotRes.Shot.Pool, snapshotRes.Shot.Era)
		if err != nil {
			return err
		}
		e.Snapshot = snapshotRes.Shot
		e.PoolUnbond = unbondRes.Unbondings
		m.Reason = core.ReasonActiveReportedEvent
		m.Content = e
		oldState[event.Type] = stafiHubXLedgerTypes.ActiveReported
		shotId = e.ShotId

	case event.Type == stafiHubXLedgerTypes.EventTypeRParamsChanged:
		if len(event.Attributes) != 5 {
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
			Denom:      denom,
			GasPrice:   rparams.RParams.GasPrice,
			EraSeconds: rparams.RParams.EraSeconds,
			LeastBond:  rparams.RParams.LeastBond,
			Offset:     rparams.RParams.Offset,
		}

		m.Reason = core.ReasonRParamsChangedEvent
		m.Content = eventRParams
	case event.Type == stafiHubXRValidatorTypes.EventTypeUpdateRValidator:
		if len(event.Attributes) != 7 {
			return ErrEventAttributeNumberUnMatch
		}
		denom := event.Attributes[0].Value
		if l.caredSymbol != core.RSymbol(denom) {
			return nil
		}
		poolAddress := event.Attributes[1].Value
		era, err := types.ParseUint(event.Attributes[2].Value)
		if err != nil {
			return err
		}

		oldAddress := event.Attributes[3].Value
		newAddress := event.Attributes[4].Value
		cycleVersion, err := types.ParseUint(event.Attributes[5].Value)
		if err != nil {
			return err
		}
		cycleNumber, err := types.ParseUint(event.Attributes[6].Value)
		if err != nil {
			return err
		}

		dealedCycle, err := l.conn.client.QueryLatestDealedCycle(denom, poolAddress)
		if err != nil {
			l.log.Warn("QueryLatestDealedCycle failed", "err", err)
			return err
		}
		// return if already dealed
		if dealedCycle.LatestDealedCycle.Number >= cycleNumber.Uint64() && dealedCycle.LatestDealedCycle.Version >= cycleVersion.Uint64() {
			return nil
		}

		eventRvalidatorUpdated := core.EventRValidatorUpdated{
			Denom:        denom,
			Era:          uint32(era.Uint64()),
			PoolAddress:  poolAddress,
			OldAddress:   oldAddress,
			NewAddress:   newAddress,
			CycleVersion: cycleVersion.Uint64(),
			CycleNumber:  cycleNumber.Uint64(),
		}
		m.Reason = core.ReasonRValidatorUpdatedEvent
		m.Content = eventRvalidatorUpdated

	default:
		return nil
	}

	l.log.Info("find event", "eventType", event.Type, "block number", blockNumber)
	err := l.submitMessage(&m)
	if err != nil {
		return err
	}

	switch m.Reason {
	// no need wait
	case core.ReasonRParamsChangedEvent:
		return nil
	case core.ReasonRValidatorUpdatedEvent:
		// here we wait until rvalidator update reported
		event, ok := m.Content.(core.EventRValidatorUpdated)
		if !ok {
			return fmt.Errorf("cast to EventRValidatorUpdated failed, event: %+v", event)
		}
		for {
			dealedCycle, err := l.conn.client.QueryLatestDealedCycle(event.Denom, event.PoolAddress)
			if err != nil {
				l.log.Warn("QueryLatestDealedCycle failed", "err", err)
				time.Sleep(BlockRetryInterval)
				continue
			}
			if !(dealedCycle.LatestDealedCycle.Number >= event.CycleNumber && dealedCycle.LatestDealedCycle.Version >= event.CycleVersion) {
				time.Sleep(BlockRetryInterval)
				continue
			}
			break
		}
	default:
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
	}

	return nil
}
