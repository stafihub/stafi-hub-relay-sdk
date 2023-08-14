package chain

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	cosMath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/sirupsen/logrus"
	"github.com/stafihub/rtoken-relay-core/common/core"
	"github.com/stafihub/stafi-hub-relay-sdk/client"
	stafiHubXLedgerTypes "github.com/stafihub/stafihub/x/ledger/types"
	stafiHubXRValidatorTypes "github.com/stafihub/stafihub/x/rvalidator/types"
)

const maxUint32 = math.MaxUint32

var (
	ErrEventAttributeNumberUnMatch = errors.New("ErrEventAttributeNumberTooFew")

	SimulateBondReportedEvent = types.StringEvent{
		Type: stafiHubXLedgerTypes.EventTypeBondReported,
		Attributes: []types.Attribute{
			{Key: stafiHubXLedgerTypes.AttributeKeyDenom, Value: "uratom"},
			{Key: stafiHubXLedgerTypes.AttributeKeyShotId, Value: "710f75313464e035f1156c115ecc1ea5b8e1cf15e5070b7198fa211273388996"}},
	}
)

func (l *Listener) processBlockEvents(currentBlock int64) error {
	if currentBlock%100 == 0 {
		l.log.Debug("processEvents", "blockNum", currentBlock)
	}

	// for dragonberry upgrade
	if currentBlock == 900501 {
		chainId, err := l.conn.client.GetChainId()
		if err != nil {
			return err
		}
		// should deal this only on mainnet
		if chainId == "stafihub-1" {
			return l.processStringEvents(SimulateBondReportedEvent, currentBlock)
		}
	}

	results, err := l.conn.client.GetBlockResults(currentBlock)
	if err != nil {
		return fmt.Errorf("client.GetBlockResults failed: %s", err)
	}

	for _, txResult := range results.TxsResults {
		for _, e := range txResult.Events {
			stringEvent, err := client.ParseBase64Event(e)
			if err != nil {
				return err
			}
			err = l.processStringEvents(stringEvent, currentBlock)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *Listener) processStringEvents(event types.StringEvent, blockNumber int64) error {
	l.log.Debug("processStringEvents", "event", event)

	msgs := make([]core.Message, 0)
	oldState := make(map[string]stafiHubXLedgerTypes.PoolBondState) // shotId => bondstate
	shotIds := make([]string, 0)

	switch event.Type {
	case stafiHubXLedgerTypes.EventTypeEraPoolUpdated:
		if len(event.Attributes)%4 != 0 {
			return ErrEventAttributeNumberUnMatch
		}
		for i := 0; i < len(event.Attributes)/4; i++ {
			lastEra, err := strconv.Atoi(event.Attributes[4*i+1].Value)
			if err != nil {
				return err
			}
			if int64(lastEra) > int64(maxUint32) {
				return fmt.Errorf("last era overflow %d", lastEra)
			}
			currentEra, err := strconv.Atoi(event.Attributes[4*i+2].Value)
			if err != nil {
				return err
			}
			if int64(currentEra) > int64(maxUint32) {
				return fmt.Errorf("current era overflow %d", currentEra)
			}
			e := core.EventEraPoolUpdated{
				Denom:      event.Attributes[4*i+0].Value,
				LastEra:    uint32(lastEra),
				CurrentEra: uint32(currentEra),
				ShotId:     event.Attributes[4*i+3].Value,
			}
			if l.caredSymbol != core.RSymbol(e.Denom) {
				continue
			}
			chainEra, err := l.conn.client.QueryChainEra(string(l.caredSymbol))
			if err != nil {
				return err
			}
			// return if already dealed
			if chainEra.GetEra() != e.CurrentEra {
				continue
			}

			snapshotRes, err := l.conn.client.QuerySnapshot(e.ShotId)
			if err != nil {
				return err
			}
			// return if already dealed
			if snapshotRes.Shot.BondState != stafiHubXLedgerTypes.EraUpdated {
				continue
			}
			e.Snapshot = snapshotRes.Shot

			m := core.Message{
				Source:      l.symbol,
				Destination: l.caredSymbol,
			}
			m.Reason = core.ReasonEraPoolUpdatedEvent
			m.Content = e

			msgs = append(msgs, m)
			oldState[e.ShotId] = stafiHubXLedgerTypes.EraUpdated
			shotIds = append(shotIds, e.ShotId)
		}
	case stafiHubXLedgerTypes.EventTypeBondReported:
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

		m := core.Message{
			Source:      l.symbol,
			Destination: l.caredSymbol,
		}
		m.Reason = core.ReasonBondReportedEvent
		m.Content = e

		msgs = append(msgs, m)
		oldState[e.ShotId] = stafiHubXLedgerTypes.BondReported
		shotIds = append(shotIds, e.ShotId)

	case stafiHubXLedgerTypes.EventTypeActiveReported:
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
		m := core.Message{
			Source:      l.symbol,
			Destination: l.caredSymbol,
		}
		m.Reason = core.ReasonActiveReportedEvent
		m.Content = e

		msgs = append(msgs, m)
		oldState[e.ShotId] = stafiHubXLedgerTypes.ActiveReported
		shotIds = append(shotIds, e.ShotId)

	case stafiHubXLedgerTypes.EventTypeRParamsChanged:
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
			Denom:      denom,
			GasPrice:   rparams.RParams.GasPrice,
			EraSeconds: rparams.RParams.EraSeconds,
			LeastBond:  rparams.RParams.LeastBond,
			Offset:     rparams.RParams.Offset,
		}

		m := core.Message{
			Source:      l.symbol,
			Destination: l.caredSymbol,
		}
		m.Reason = core.ReasonRParamsChangedEvent
		m.Content = eventRParams

		msgs = append(msgs, m)

	case stafiHubXRValidatorTypes.EventTypeUpdateRValidator:
		if len(event.Attributes) != 8 {
			return ErrEventAttributeNumberUnMatch
		}
		denom := event.Attributes[0].Value
		if l.caredSymbol != core.RSymbol(denom) {
			return nil
		}
		poolAddress := event.Attributes[1].Value
		era, err := cosMath.ParseUint(event.Attributes[2].Value)
		if err != nil {
			return err
		}

		oldAddress := event.Attributes[3].Value
		newAddress := event.Attributes[4].Value
		cycleVersion, err := cosMath.ParseUint(event.Attributes[5].Value)
		if err != nil {
			return err
		}
		cycleNumber, err := cosMath.ParseUint(event.Attributes[6].Value)
		if err != nil {
			return err
		}
		cycleSeconds, err := cosMath.ParseUint(event.Attributes[7].Value)
		if err != nil {
			return err
		}

		dealedCycle, err := l.conn.client.QueryLatestDealedCycle(denom, poolAddress)
		if err != nil {
			if !strings.Contains(err.Error(), "NotFound") {
				l.log.Warn("QueryLatestDealedCycle failed", "err", err)
				return err
			}
		} else {
			// return if already dealed
			if dealedCycle.LatestDealedCycle.Number >= cycleNumber.Uint64() && dealedCycle.LatestDealedCycle.Version >= cycleVersion.Uint64() {
				return nil
			}
		}
		resultBlock, err := l.conn.client.QueryBlock(blockNumber)
		if err != nil {
			return err
		}

		eventRvalidatorUpdated := core.EventRValidatorUpdated{
			Denom:          denom,
			Era:            uint32(era.Uint64()),
			PoolAddress:    poolAddress,
			OldAddress:     oldAddress,
			NewAddress:     newAddress,
			CycleVersion:   cycleVersion.Uint64(),
			CycleNumber:    cycleNumber.Uint64(),
			CycleSeconds:   cycleSeconds.Uint64(),
			BlockTimestamp: resultBlock.Block.Time.Unix(),
		}

		m := core.Message{
			Source:      l.symbol,
			Destination: l.caredSymbol,
		}

		m.Reason = core.ReasonRValidatorUpdatedEvent
		m.Content = eventRvalidatorUpdated

		msgs = append(msgs, m)
	case stafiHubXRValidatorTypes.EventTypeAddRValidator:
		if len(event.Attributes) != 4 {
			return ErrEventAttributeNumberUnMatch
		}
		denom := event.Attributes[0].Value
		if l.caredSymbol != core.RSymbol(denom) {
			return nil
		}
		poolAddress := event.Attributes[1].Value
		era, err := cosMath.ParseUint(event.Attributes[2].Value)
		if err != nil {
			return err
		}

		addedAddress := event.Attributes[3].Value

		eventRvalidatorAdded := core.EventRValidatorAdded{
			Denom:        denom,
			Era:          uint32(era.Uint64()),
			PoolAddress:  poolAddress,
			AddedAddress: addedAddress,
		}

		m := core.Message{
			Source:      l.symbol,
			Destination: l.caredSymbol,
		}
		m.Reason = core.ReasonRValidatorAddedEvent
		m.Content = eventRvalidatorAdded

		msgs = append(msgs, m)
	case stafiHubXLedgerTypes.EventTypeInitPool:
		if len(event.Attributes) != 2 {
			return ErrEventAttributeNumberUnMatch
		}
		denom := event.Attributes[0].Value
		if l.caredSymbol != core.RSymbol(denom) {
			return nil
		}
		poolAddress := event.Attributes[1].Value

		icaPoolList, err := l.conn.client.QueryIcaPoolList(denom)
		if err != nil {
			return err
		}

		withdrawalAddr := ""
		hostChannelId := ""
		for _, icaPool := range icaPoolList.IcaPoolList {
			if icaPool.DelegationAccount.Address == poolAddress {
				withdrawalAddr = icaPool.WithdrawalAccount.Address
				hostChannelId = icaPool.DelegationAccount.HostChannelId
				break
			}
		}
		if len(withdrawalAddr) == 0 || len(hostChannelId) == 0 {
			logrus.Info("init pool but not ica pool: ", poolAddress)
			return nil
		}

		// wait until get rvalidators
		var vals []string
		retry := BlockRetryLimit
		for {
			if retry <= 0 {
				return fmt.Errorf("QueryRValidatorList reach retry limit: denom %s pool %s", denom, poolAddress)
			}

			validators, err := l.conn.client.QueryRValidatorList(denom, poolAddress)
			if err != nil {
				return err
			}
			if len(validators.RValidatorList) == 0 {
				time.Sleep(BlockRetryInterval)
				retry++
				continue
			}
			vals = validators.RValidatorList
			break
		}

		eventInitPool := core.EventInitPool{
			Denom:             denom,
			PoolAddress:       poolAddress,
			WithdrawalAddress: withdrawalAddr,
			HostChannelId:     hostChannelId,
			Validators:        vals,
		}

		m := core.Message{
			Source:      l.symbol,
			Destination: l.caredSymbol,
		}
		m.Reason = core.ReasonInitPoolEvent
		m.Content = eventInitPool

		msgs = append(msgs, m)
	case stafiHubXLedgerTypes.EventTypeRemovePool:
		if len(event.Attributes) != 2 {
			return ErrEventAttributeNumberUnMatch
		}
		denom := event.Attributes[0].Value
		if l.caredSymbol != core.RSymbol(denom) {
			return nil
		}
		poolAddress := event.Attributes[1].Value

		eventRemovePool := core.EventRemovePool{
			Denom:       denom,
			PoolAddress: poolAddress,
		}

		m := core.Message{
			Source:      l.symbol,
			Destination: l.caredSymbol,
		}
		m.Reason = core.ReasonRemovePoolEvent
		m.Content = eventRemovePool

		msgs = append(msgs, m)

	default:
		return nil
	}

	l.log.Info("find event", "eventType", event.Type, "block number", blockNumber, "msgs", msgs)
	for index, msg := range msgs {

		err := l.submitMessage(&msg)
		if err != nil {
			return err
		}

		switch msg.Reason {
		case core.ReasonRParamsChangedEvent, core.ReasonRValidatorAddedEvent, core.ReasonInitPoolEvent, core.ReasonRemovePoolEvent:
			// no need wait, we will get latest state when restart
			return nil
		case core.ReasonRValidatorUpdatedEvent:
			// events of rvalidator updated
			// here we wait until rvalidator update reported
			// so we can continuely process this event when restart
			event, ok := msg.Content.(core.EventRValidatorUpdated)
			if !ok {
				return fmt.Errorf("cast to EventRValidatorUpdated failed, event: %+v", event)
			}
			for {
				dealedCycle, err := l.conn.client.QueryLatestDealedCycle(event.Denom, event.PoolAddress)
				if err != nil {
					l.log.Warn("QueryLatestDealedCycle failed will retry", "err", err)
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
			// events of era dealing
			// here we wait until snapshot's bondstate change to another
			// so we can continuely process this event when restart
			shotId := shotIds[index]
			for {
				snapshotRes, err := l.conn.client.QuerySnapshot(shotId)
				if err != nil {
					l.log.Warn("QuerySnapshot failed will retry", "err", err)
					time.Sleep(BlockRetryInterval)
					continue
				}
				if snapshotRes.GetShot().BondState == oldState[shotId] {
					time.Sleep(BlockRetryInterval)
					continue
				}
				break
			}
		}
	}
	return nil
}
