package chain

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/stafiprotocol/rtoken-relay-core/common/core"
	stafiHubXLedgerTypes "github.com/stafiprotocol/stafihub/x/ledger/types"
)

const maxUint32 = ^uint32(0)
const maxInt32 = int32(maxUint32 >> 1)

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
	switch {
	case event.Type == stafiHubXLedgerTypes.EventTypeEraPoolUpdated:
		if len(event.Attributes) != 5 {
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
			Denom:       event.Attributes[0].Value,
			LastEra:     uint32(lastEra),
			CurrentEra:  uint32(currentEra),
			ShotId:      event.Attributes[3].Value,
			LasterVoter: event.Attributes[4].Value,
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

		shotId, err := hex.DecodeString(e.ShotId)
		if err != nil {
			return err
		}
		snapshotRes, err := l.conn.client.QuerySnapshot(shotId)
		if err != nil {
			return err
		}
		if snapshotRes.Shot.BondState != stafiHubXLedgerTypes.EraUpdated {
			return nil
		}
		e.Snapshot = snapshotRes.Shot
		m.Reason = core.ReasonEraPoolUpdatedEvent
		m.Content = e

	case event.Type == stafiHubXLedgerTypes.EventTypeBondReported:
		if len(event.Attributes) != 3 {
			return ErrEventAttributeNumberUnMatch
		}

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
		snapshotRes, err := l.conn.client.QuerySnapshot(shotId)
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
	case event.Type == stafiHubXLedgerTypes.EventTypeActiveReported:
		if len(event.Attributes) != 3 {
			return ErrEventAttributeNumberUnMatch
		}

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
		snapshotRes, err := l.conn.client.QuerySnapshot(shotId)
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
	case event.Type == stafiHubXLedgerTypes.EventTypeWithdrawReported:
		return nil
		// if len(event.Attributes) != 3 {
		// 	return ErrEventAttributeNumberUnMatch
		// }

		// e := core.EventWithdrawReported{
		// 	Denom:       event.Attributes[0].Value,
		// 	ShotId:      event.Attributes[1].Value,
		// 	LasterVoter: event.Attributes[2].Value,
		// }
		// if l.caredSymbol != core.RSymbol(e.Denom) {
		// 	return nil
		// }
		// shotId, err := hex.DecodeString(e.ShotId)
		// if err != nil {
		// 	return err
		// }
		// snapshotRes, err := l.conn.client.QuerySnapshot(shotId)
		// if err != nil {
		// 	return err
		// }
		// unbondRes, err := l.conn.client.QueryPoolUnbond(e.Denom, snapshotRes.Shot.Pool, snapshotRes.Shot.Era)
		// if err != nil {
		// 	return err
		// }

		// e.Snapshot = snapshotRes.Shot
		// e.PoolUnbond = unbondRes.Unbond
		// m.Reason = core.ReasonWithdrawReportedEvent
		// m.Content = e
	case event.Type == stafiHubXLedgerTypes.EventTypeTransferReported:
		return nil
		// if len(event.Attributes) != 3 {
		// 	return ErrEventAttributeNumberUnMatch
		// }

		// e := core.EventTransferReported{
		// 	Denom:       event.Attributes[0].Value,
		// 	ShotId:      event.Attributes[1].Value,
		// 	LasterVoter: event.Attributes[2].Value,
		// }
		// if l.caredSymbol != core.RSymbol(e.Denom) {
		// 	return nil
		// }
		// m.Reason = core.ReasonTransferReportedEvent
		// m.Content = e
	case event.Type == stafiHubXLedgerTypes.EventTypeSignatureEnough:
		if len(event.Attributes) != 5 {
			return ErrEventAttributeNumberUnMatch
		}

		era, err := strconv.Atoi(event.Attributes[1].Value)
		if err != nil {
			return err
		}
		if int64(era) > int64(maxInt32) {
			return fmt.Errorf("era overflow %d", era)
		}

		txType := event.Attributes[3].Value
		useTxType := stafiHubXLedgerTypes.OriginalTxType_value[txType]

		proposalId, err := hex.DecodeString(event.Attributes[4].Value)
		if err != nil {
			return err
		}

		e := core.EventSignatureEnough{
			Denom:      event.Attributes[0].Value,
			Era:        uint32(era),
			Pool:       event.Attributes[2].Value,
			TxType:     stafiHubXLedgerTypes.OriginalTxType(useTxType),
			ProposalId: proposalId,
			Signatures: [][]byte{},
		}
		if l.caredSymbol != core.RSymbol(e.Denom) {
			return nil
		}
		poolDetail, err := l.conn.client.QueryPoolDetail(e.Denom, e.Pool)
		if err != nil {
			return err
		}
		e.Threshold = poolDetail.Detail.Threshold

		signature, err := l.conn.client.QuerySignature(e.Denom, e.Pool, e.Era, e.TxType, e.ProposalId)
		if err != nil {
			return err
		}
		for _, v := range signature.Signature.GetSigs() {
			sigBts, err := hex.DecodeString(v)
			if err != nil {
				return err
			}
			e.Signatures = append(e.Signatures, sigBts)
		}

		m.Reason = core.ReasonSignatureEnoughEvent
		m.Content = e

	default:
		return nil
	}
	l.log.Info("find event", "msg", m, "block number", blockNumber)
	return l.submitMessage(&m)
}
