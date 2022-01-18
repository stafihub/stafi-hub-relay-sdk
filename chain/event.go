package chain

import (
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
				//commonly the number of events we will handle in a block is less or equal to 1
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
		e := EventEraPoolUpdated{
			Denom:       event.Attributes[0].Value,
			LastEra:     event.Attributes[1].Value,
			CurrentEra:  event.Attributes[2].Value,
			ShotId:      event.Attributes[3].Value,
			LasterVoter: event.Attributes[4].Value,
		}
		if l.caredSymbol != core.RSymbol(e.Denom) {
			return nil
		}
		m.Reason = core.ReasonEraPoolUpdatedEvent
		m.Content = e
	case event.Type == stafiHubXLedgerTypes.EventTypeBondReported:
		e := EventBondReported{
			Denom:       event.Attributes[0].Value,
			ShotId:      event.Attributes[1].Value,
			LasterVoter: event.Attributes[2].Value,
		}
		if l.caredSymbol != core.RSymbol(e.Denom) {
			return nil
		}
		m.Reason = core.ReasonBondReportedEvent
		m.Content = e
	case event.Type == stafiHubXLedgerTypes.EventTypeActiveReported:
		e := EventActiveReported{
			Denom:       event.Attributes[0].Value,
			ShotId:      event.Attributes[1].Value,
			LasterVoter: event.Attributes[2].Value,
		}
		if l.caredSymbol != core.RSymbol(e.Denom) {
			return nil
		}
		m.Reason = core.ReasonActiveReportedEvent
		m.Content = e
	case event.Type == stafiHubXLedgerTypes.EventTypeWithdrawReported:
		e := EventWithdrawReported{
			Denom:       event.Attributes[0].Value,
			ShotId:      event.Attributes[1].Value,
			LasterVoter: event.Attributes[2].Value,
		}
		if l.caredSymbol != core.RSymbol(e.Denom) {
			return nil
		}
		m.Reason = core.ReasonWithdrawReportedEvent
		m.Content = e
	case event.Type == stafiHubXLedgerTypes.EventTypeTransferReported:
		e := EventTransferReported{
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

type EventEraPoolUpdated struct {
	Denom       string
	LastEra     string
	CurrentEra  string
	ShotId      string
	LasterVoter string
	Snapshot    stafiHubXLedgerTypes.BondSnapshot
}

type EventBondReported struct {
	Denom       string
	ShotId      string
	LasterVoter string
	Snapshot    stafiHubXLedgerTypes.BondSnapshot
}

type EventActiveReported struct {
	Denom       string
	ShotId      string
	LasterVoter string
	Snapshot    stafiHubXLedgerTypes.BondSnapshot
}

type EventWithdrawReported struct {
	Denom       string
	ShotId      string
	LasterVoter string
	Snapshot    stafiHubXLedgerTypes.BondSnapshot
	PoolUnbond  stafiHubXLedgerTypes.PoolUnbond
}

type EventTransferReported struct {
	Denom       string
	ShotId      string
	LasterVoter string
}
