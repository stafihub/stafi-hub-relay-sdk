package chain

import (
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/stafiprotocol/rtoken-relay-core/core"
	stafiHubXLedgerTypes "github.com/stafiprotocol/stafihub/x/ledger/types"
)

func (l *Listener) processBlockEvents(currentBlock int64) error {
	if currentBlock%100 == 0 {
		l.log.Debug("processEvents", "blockNum", currentBlock)
	}

	txs, err := l.conn.client.GetBlockTxs(currentBlock)
	if err != nil {
		return err
	}
	for _, tx := range txs {
		for _, log := range tx.Logs {
			for _, event := range log.Events {
				err := l.processStringEvents(event)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (l *Listener) processStringEvents(event types.StringEvent) error {
	m := core.Message{
		Source:      l.symbol,
		Destination: l.caredSymbol,
	}
	switch {
	case event.Type == stafiHubXLedgerTypes.EventTypeEraPoolUpdated:
		e := core.EventEraPoolUpdated{
			Denom:       event.Attributes[0].Value,
			LastEra:     event.Attributes[1].Value,
			CurrentEra:  event.Attributes[2].Value,
			ShotId:      event.Attributes[3].Value,
			LasterVoter: event.Attributes[4].Value,
		}
		if l.caredSymbol != core.RSymbol(e.Denom) {
			return nil
		}
		shotId, err := hex.DecodeString(e.ShotId)
		if err != nil {
			return err
		}
		snapshotRes, err := l.conn.QuerySnapshot(shotId)
		if err != nil {
			return err
		}
		e.Snapshot = snapshotRes.Shot
		m.Reason = core.ReasonEraPoolUpdatedEvent
		m.Content = e

	case event.Type == stafiHubXLedgerTypes.EventTypeBondReported:
		e := core.EventBondReported{
			Denom:       event.Attributes[0].Value,
			ShotId:      event.Attributes[1].Value,
			LasterVoter: event.Attributes[2].Value,
		}
		if l.caredSymbol != core.RSymbol(e.Denom) {
			return nil
		}
		shotId, err := hex.DecodeString(e.ShotId)
		if err != nil {
			return err
		}
		snapshotRes, err := l.conn.QuerySnapshot(shotId)
		if err != nil {
			return err
		}
		e.Snapshot = snapshotRes.Shot
		m.Reason = core.ReasonBondReportedEvent
		m.Content = e
	case event.Type == stafiHubXLedgerTypes.EventTypeActiveReported:
		e := core.EventActiveReported{
			Denom:       event.Attributes[0].Value,
			ShotId:      event.Attributes[1].Value,
			LasterVoter: event.Attributes[2].Value,
		}
		if l.caredSymbol != core.RSymbol(e.Denom) {
			return nil
		}
		shotId, err := hex.DecodeString(e.ShotId)
		if err != nil {
			return err
		}
		snapshotRes, err := l.conn.QuerySnapshot(shotId)
		if err != nil {
			return err
		}
		e.Snapshot = snapshotRes.Shot
		m.Reason = core.ReasonActiveReportedEvent
		m.Content = e
	case event.Type == stafiHubXLedgerTypes.EventTypeWithdrawReported:
		e := core.EventWithdrawReported{
			Denom:       event.Attributes[0].Value,
			ShotId:      event.Attributes[1].Value,
			LasterVoter: event.Attributes[2].Value,
		}
		if l.caredSymbol != core.RSymbol(e.Denom) {
			return nil
		}
		shotId, err := hex.DecodeString(e.ShotId)
		if err != nil {
			return err
		}
		snapshotRes, err := l.conn.QuerySnapshot(shotId)
		if err != nil {
			return err
		}
		unbondRes, err := l.conn.QueryPoolUnbond(e.Denom, snapshotRes.Shot.Pool, snapshotRes.Shot.Era)
		if err != nil {
			return err
		}

		e.Snapshot = snapshotRes.Shot
		e.PoolUnbond = unbondRes.Unbond
		m.Reason = core.ReasonWithdrawReportedEvent
		m.Content = e
	case event.Type == stafiHubXLedgerTypes.EventTypeTransferReported:
		e := core.EventTransferReported{
			Denom:       event.Attributes[0].Value,
			ShotId:      event.Attributes[1].Value,
			LasterVoter: event.Attributes[2].Value,
		}
		if l.caredSymbol != core.RSymbol(e.Denom) {
			return nil
		}
		m.Reason = core.ReasonTransferReportedEvent
		m.Content = e
	default:
		return fmt.Errorf("not support event type: %s", event.Type)
	}
	return l.submitMessage(&m)
}
