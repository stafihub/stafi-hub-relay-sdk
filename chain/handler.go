package chain

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stafihub/rtoken-relay-core/common/core"
	"github.com/stafihub/rtoken-relay-core/common/log"
	hubClient "github.com/stafihub/stafi-hub-relay-sdk/client"
	"github.com/stafihub/stafihub/utils"
	stafiHubXLedgerTypes "github.com/stafihub/stafihub/x/ledger/types"
	stafiHubXRelayersTypes "github.com/stafihub/stafihub/x/relayers/types"
	stafiHubXRValidatorTypes "github.com/stafihub/stafihub/x/rvalidator/types"
	stafiHubXRVoteTypes "github.com/stafihub/stafihub/x/rvote/types"
	"google.golang.org/grpc/codes"
)

const msgLimit = 512

var oneDec = utils.NewDecFromBigInt(big.NewInt(1))

type Handler struct {
	conn            *Connection
	router          *core.Router
	msgChan         chan *core.Message
	priorityMsgChan chan *core.Message
	log             log.Logger
	stopChan        <-chan struct{}
	sysErrChan      chan<- error
}

func NewHandler(conn *Connection, log log.Logger, stopChan <-chan struct{}, sysErrChan chan<- error) *Handler {
	return &Handler{
		conn:            conn,
		msgChan:         make(chan *core.Message, msgLimit),
		priorityMsgChan: make(chan *core.Message, msgLimit),
		log:             log,
		stopChan:        stopChan,
		sysErrChan:      sysErrChan,
	}
}

func (w *Handler) setRouter(r *core.Router) {
	w.router = r
}

func (w *Handler) start() error {
	go w.msgHandler()
	return nil
}

// resolve msg from other chains
func (w *Handler) HandleMessage(m *core.Message) {
	// deal get msg
	switch m.Reason {
	case core.ReasonGetPools:
		go w.handleGetPools(m)
		return
	case core.ReasonGetSignatures:
		go w.handleGetSignatures(m)
		return
	case core.ReasonGetBondRecord:
		go w.handleGetBondRecord(m)
		return
	case core.ReasonGetInterchainTxStatus:
		go w.handleGetInterchainTxStatus(m)
		return
	}
	// deal write msg
	w.queueMessage(m)
}

func (w *Handler) queueMessage(m *core.Message) {
	if m.Reason == core.ReasonSubmitSignature {
		w.priorityMsgChan <- m
	} else {
		w.msgChan <- m
	}
}

func (w *Handler) msgHandler() {
	for {
		select {
		case <-w.stopChan:
			w.log.Info("priorityMsgHandler receive stopChain, will stop")
			return
		case msg := <-w.priorityMsgChan:
			err := w.handleMessage(msg)
			if err != nil {
				w.sysErrChan <- fmt.Errorf("resolvePriorityMessage process failed.err: %s, msg: %+v", err, msg)
				return
			}
		default:
		}

		select {
		case <-w.stopChan:
			w.log.Info("msgHandler receive stopChain, will stop")
			return
		case msg := <-w.msgChan:
			err := w.handleMessage(msg)
			if err != nil {
				w.sysErrChan <- fmt.Errorf("resolveMessage process failed.err: %s, msg: %+v", err, msg)
				return
			}
		default:
		}

		if len(w.msgChan) == 0 && len(w.priorityMsgChan) == 0 {
			time.Sleep(2 * time.Second)
		}
	}
}

// resolve write msg from other chains
func (w *Handler) handleMessage(m *core.Message) error {
	switch m.Reason {
	case core.ReasonExeLiquidityBond:
		return w.handleExeLiquidityBond(m)
	case core.ReasonNewEra:
		return w.handleNewChainEra(m)
	case core.ReasonBondReport:
		return w.handleBondReport(m)
	case core.ReasonActiveReport:
		return w.handleActiveReport(m)
	case core.ReasonTransferReport:
		return w.handleTransferReport(m)
	case core.ReasonRValidatorUpdateReport:
		return w.handleRValidatorUpdateReport(m)
	case core.ReasonSubmitSignature:
		return w.handleSubmitSignature(m)
	case core.ReasonInterchainTx:
		return w.handleInterchainTx(m)
	default:
		return fmt.Errorf("message reason unsupported reason: %s", m.Reason)
	}
}

func (w *Handler) handleExeLiquidityBond(m *core.Message) error {
	w.log.Info("handleExeLiquidityBond", "m", m)
	proposal, ok := m.Content.(core.ProposalExeLiquidityBond)
	if !ok {
		return fmt.Errorf("ProposalLiquidityBond cast failed, %+v", m)
	}
	recordRes, err := w.conn.client.QueryBondRecord(proposal.Denom, proposal.Txhash)
	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		return err
	}
	if err == nil && recordRes.BondRecord.State == stafiHubXLedgerTypes.LiquidityBondStateVerifyOk {
		w.log.Warn("handleExeLiquidityBond already verifyOk, no need submitproposal", "msg", m)
		return nil
	}

	done := core.UseSdkConfigContext(hubClient.GetAccountPrefix())
	bonder, err := types.AccAddressFromBech32(proposal.Bonder)
	if err != nil {
		done()
		return err
	}
	content := stafiHubXLedgerTypes.NewExecuteBondProposal(
		w.conn.client.GetFromAddress(), proposal.Denom, bonder,
		proposal.Pool, proposal.Txhash, proposal.Amount, proposal.State)
	done()

	return w.checkAndReSendWithProposalContent("executeBondProposal", content)
}

func (w *Handler) handleNewChainEra(m *core.Message) error {
	w.log.Debug("handleNewChainEra", "m", m)
	proposal, ok := m.Content.(core.ProposalSetChainEra)
	if !ok {
		return fmt.Errorf("ProposalSetChainEra cast failed, %+v", m)
	}

	eraOnChain := uint32(0)
	chainEra, err := w.conn.client.QueryChainEra(proposal.Denom)
	if err != nil {
		if strings.Contains(err.Error(), codes.NotFound.String()) {
			eraOnChain = proposal.Era - 1
		} else {
			return err
		}
	} else {
		eraOnChain = chainEra.GetEra()
	}

	if eraOnChain >= proposal.Era {
		return nil
	}
	continuable, err := w.conn.client.QueryEraContinuable(proposal.Denom, eraOnChain)
	if err != nil {
		return err
	}
	if !continuable {
		return nil
	}

	useEra := eraOnChain + 1
	w.log.Info("will set newChainEra", "newEra", useEra, "symbol", proposal.Denom)

	done := core.UseSdkConfigContext(hubClient.GetAccountPrefix())
	content := stafiHubXLedgerTypes.NewSetChainEraProposal(w.conn.client.GetFromAddress(), proposal.Denom, useEra)
	done()

	return w.checkAndReSendWithProposalContent("setChainEraProposal", content)
}

func (w *Handler) handleBondReport(m *core.Message) error {
	w.log.Info("handleBondReport", "m", m)
	proposal, ok := m.Content.(core.ProposalBondReport)
	if !ok {
		return fmt.Errorf("ProposalBondReport cast failed, %+v", m)
	}

	done := core.UseSdkConfigContext(hubClient.GetAccountPrefix())
	content := stafiHubXLedgerTypes.NewBondReportProposal(w.conn.client.GetFromAddress(), proposal.Denom, proposal.ShotId, proposal.Action)
	done()

	return w.checkAndReSendWithProposalContent("bondReportProposal", content)
}

func (w *Handler) handleActiveReport(m *core.Message) error {
	w.log.Info("handleActiveReport", "m", m)
	proposal, ok := m.Content.(core.ProposalActiveReport)
	if !ok {
		return fmt.Errorf("ProposalActiveReport cast failed, %+v", m)
	}
	// ensure rate >=1
	rate, err := w.conn.client.QueryRate(proposal.Denom)
	if err != nil {
		return err
	}
	if rate.ExchangeRate.Value.Equal(oneDec) {
		snapshot, err := w.conn.client.QuerySnapshot(proposal.ShotId)
		if err != nil {
			return err
		}
		if snapshot.Shot.Chunk.Active.Sub(proposal.Staked).LT(types.NewIntFromUint64(1000)) {
			proposal.Staked = snapshot.Shot.Chunk.Active
		}
	}

	done := core.UseSdkConfigContext(hubClient.GetAccountPrefix())
	content := stafiHubXLedgerTypes.NewActiveReportProposal(w.conn.client.GetFromAddress(), proposal.Denom, proposal.ShotId, proposal.Staked, proposal.Unstaked)
	done()

	return w.checkAndReSendWithProposalContent("activeReportProposal", content)
}

func (w *Handler) handleTransferReport(m *core.Message) error {
	w.log.Info("handleTransferReport", "m", m)
	proposal, ok := m.Content.(core.ProposalTransferReport)
	if !ok {
		return fmt.Errorf("ProposalTransferReport cast failed, %+v", m)
	}

	done := core.UseSdkConfigContext(hubClient.GetAccountPrefix())
	content := stafiHubXLedgerTypes.NewTransferReportProposal(w.conn.client.GetFromAddress(), proposal.Denom, proposal.ShotId)
	done()

	return w.checkAndReSendWithProposalContent("transferReportProposal", content)
}

func (w *Handler) handleInterchainTx(m *core.Message) error {
	w.log.Info("handleInterchainTx", "m", m)
	proposal, ok := m.Content.(core.ProposalInterchainTx)
	if !ok {
		return fmt.Errorf("ProposalInterchainTx cast failed, %+v", m)
	}

	done := core.UseSdkConfigContext(hubClient.GetAccountPrefix())
	content, err := stafiHubXLedgerTypes.NewInterchainTxProposal(
		w.conn.client.GetFromAddress(),
		proposal.Denom,
		proposal.Pool,
		proposal.Era,
		proposal.TxType,
		proposal.Factor,
		proposal.Msgs)
	if err != nil {
		done()
		return err
	}
	done()

	return w.checkAndReSendWithProposalContent("interchainTxProposal", content)
}

func (w *Handler) handleRValidatorUpdateReport(m *core.Message) error {
	w.log.Info("handleRValidatorUpdateReport", "m", m)
	proposal, ok := m.Content.(core.ProposalRValidatorUpdateReport)
	if !ok {
		return fmt.Errorf("ProposalRValidatorUpdateReport cast failed, %+v", m)
	}

	done := core.UseSdkConfigContext(hubClient.GetAccountPrefix())
	content := stafiHubXRValidatorTypes.NewUpdateRValidatorReportProposal(
		w.conn.client.GetFromAddress().String(),
		proposal.Denom,
		proposal.PoolAddress,
		&stafiHubXRValidatorTypes.Cycle{
			Denom:       proposal.Denom,
			PoolAddress: proposal.PoolAddress,
			Version:     proposal.CycleVersion,
			Number:      proposal.CycleNumber,
		}, proposal.Status)
	done()

	return w.checkAndReSendWithProposalContent("ProposalRValidatorUpdateReport", content)
}

func (w *Handler) handleSubmitSignature(m *core.Message) error {
	w.log.Debug("handleSubmitSignature", "m", m)
	proposal, ok := m.Content.(core.ParamSubmitSignature)
	if !ok {
		return fmt.Errorf("ProposalSubmitSignature cast failed, %+v", m)
	}

	done := core.UseSdkConfigContext(hubClient.GetAccountPrefix())
	msg := stafiHubXLedgerTypes.NewMsgSubmitSignature(w.conn.client.GetFromAddress().String(), proposal.Denom, proposal.Era, proposal.Pool, proposal.TxType, proposal.PropId, proposal.Signature)
	done()

	return w.checkAndReSendWithSubmitSignature("submitSignature", msg)
}

func (w *Handler) handleGetPools(m *core.Message) error {
	w.log.Info("handleGetPools", "m", m)
	getPools, ok := m.Content.(core.ParamGetPools)
	if !ok {
		return fmt.Errorf("GetPools cast failed, %+v", m)
	}
	pools, err := w.conn.client.QueryPools(getPools.Denom)
	if err != nil {
		getPools.Pools <- []string{}
		return nil
	}
	getPools.Pools <- pools.GetAddrs()

	w.log.Info("getPools", "pools", pools.GetAddrs())
	return nil
}

func (w *Handler) handleGetSignatures(m *core.Message) error {
	w.log.Debug("handleGetSignature", "m", m)
	getSiganture, ok := m.Content.(core.ParamGetSignatures)
	if !ok {
		return fmt.Errorf("GetSignature cast failed, %+v", m)
	}
	sigs, err := w.conn.client.QuerySignature(getSiganture.Denom, getSiganture.Pool, getSiganture.Era, getSiganture.TxType, getSiganture.PropId)
	if err != nil {
		getSiganture.Sigs <- []string{}
		return nil
	}
	getSiganture.Sigs <- sigs.Signature.Sigs

	w.log.Debug("getSignatures", "sigs", sigs.Signature.Sigs)
	return nil
}

func (w *Handler) handleGetBondRecord(m *core.Message) error {
	w.log.Debug("handleGetBondRecord", "m", m)
	getBondRecord, ok := m.Content.(core.ParamGetBondRecord)
	if !ok {
		return fmt.Errorf("GetBondRecord cast failed, %+v", m)
	}
	bondRecord, err := w.conn.client.QueryBondRecord(getBondRecord.Denom, getBondRecord.TxHash)
	if err != nil {
		getBondRecord.BondRecord <- stafiHubXLedgerTypes.BondRecord{}
		return nil
	}
	getBondRecord.BondRecord <- bondRecord.BondRecord

	w.log.Debug("getBondRecord", "bondRecord", bondRecord.BondRecord)
	return nil
}

func (w *Handler) handleGetInterchainTxStatus(m *core.Message) error {
	w.log.Debug("handleGetInterchainTxStatus", "m", m)
	getInterchainTxStatus, ok := m.Content.(core.ParamGetInterchainTxStatus)
	if !ok {
		return fmt.Errorf("ParamGetInterchainTxStatus cast failed, %+v", m)
	}
	statusRes, err := w.conn.client.QueryInterchainTxStatus(getInterchainTxStatus.PropId)
	if err != nil {
		getInterchainTxStatus.Status <- stafiHubXLedgerTypes.InterchainTxStatusUnspecified
		return nil
	}
	getInterchainTxStatus.Status <- statusRes.InterchainTxStatus

	w.log.Debug("getInterchainTxStatus", "status", statusRes.InterchainTxStatus)
	return nil
}

func (h *Handler) checkAndReSendWithSubmitSignature(typeStr string, sigMsg *stafiHubXLedgerTypes.MsgSubmitSignature) error {
	txHashStr, _, err := h.conn.client.SubmitSignature(sigMsg)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), stafiHubXLedgerTypes.ErrSignatureRepeated.Error()):
			h.log.Info("no need send, already submit signature", "txHash", txHashStr, "type", typeStr)
			return nil
		}
		return err
	}

	retry := BlockRetryLimit
	for {
		var err error
		var res *types.TxResponse
		if retry <= 0 {
			h.log.Error(
				"checkAndReSendWithSubmitSignature QueryTxByHash, reach retry limit.",
				"tx hash", txHashStr,
				"err", err)
			return fmt.Errorf("checkAndReSendWithSubmitSignature QueryTxByHash reach retry limit, tx hash: %s", txHashStr)
		}
		//check on chain
		res, err = h.conn.client.QueryTxByHash(txHashStr)
		if err != nil || res.Empty() || res.Height == 0 {
			if res != nil {
				h.log.Debug(fmt.Sprintf(
					"checkAndReSendWithSubmitSignature QueryTxByHash, tx failed. will query after %f second",
					BlockRetryInterval.Seconds()),
					"tx hash", txHashStr,
					"res.log", res.RawLog,
					"res.code", res.Code)
			} else {
				h.log.Debug(fmt.Sprintf(
					"checkAndReSendWithSubmitSignature QueryTxByHash failed. will query after %f second",
					BlockRetryInterval.Seconds()),
					"tx hash", txHashStr,
					"err", err)
			}

			time.Sleep(BlockRetryInterval)
			retry--
			continue
		}

		if res.Code != 0 {
			switch {
			case strings.Contains(res.RawLog, stafiHubXLedgerTypes.ErrSignatureRepeated.Error()):
				h.log.Info("no need send, already submit signature", "txHash", txHashStr, "type", typeStr)
				return nil
				// resend case
			case strings.Contains(res.RawLog, errors.ErrOutOfGas.Error()):
				return h.checkAndReSendWithSubmitSignature(txHashStr, sigMsg)
			default:
				return fmt.Errorf("tx failed, txHash: %s, rawlog: %s", txHashStr, res.RawLog)
			}
		}

		break
	}

	h.log.Info("checkAndReSendWithSubmitSignature success", "txHash", txHashStr, "type", typeStr)
	return nil
}

func (h *Handler) checkAndReSendWithProposalContent(typeStr string, content stafiHubXRVoteTypes.Content) error {
	txHashStr, _, err := h.conn.client.SubmitProposal(content)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), stafiHubXRelayersTypes.ErrAlreadyVoted.Error()):
			h.log.Info("no need send, already voted", "txHash", txHashStr, "type", typeStr)
			return nil
		case strings.Contains(err.Error(), stafiHubXRVoteTypes.ErrProposalAlreadyApproved.Error()):
			h.log.Info("no need send, already approved", "txHash", txHashStr, "type", typeStr)
			return nil
		case strings.Contains(err.Error(), stafiHubXRVoteTypes.ErrProposalAlreadyExpired.Error()):
			h.log.Info("no need send, already expired", "txHash", txHashStr, "type", typeStr)
			return nil
		case strings.Contains(err.Error(), stafiHubXLedgerTypes.ErrEraNotContinuable.Error()):
			h.log.Info("no need send, already update new era", "txHash", txHashStr, "type", typeStr)
			return nil

		// resend case:
		case strings.Contains(err.Error(), errors.ErrWrongSequence.Error()):
			return h.checkAndReSendWithProposalContent(txHashStr, content)
		}

		return err
	}

	retry := BlockRetryLimit
	var res *types.TxResponse
	for {
		if retry <= 0 {
			h.log.Error("checkAndReSendWithProposalContent QueryTxByHash, reach retry limit.",
				"tx hash", txHashStr,
				"err", err)
			return fmt.Errorf("checkAndReSendWithProposalContent QueryTxByHash reach retry limit, tx hash: %s,err: %s", txHashStr, err)
		}

		//check on chain
		res, err = h.conn.client.QueryTxByHash(txHashStr)
		if err != nil || res.Empty() || res.Height == 0 {
			if res != nil {
				h.log.Debug(fmt.Sprintf(
					"checkAndReSendWithProposalContent QueryTxByHash, tx failed. will query after %f second",
					BlockRetryInterval.Seconds()),
					"tx hash", txHashStr,
					"res.log", res.RawLog,
					"res.code", res.Code)
			} else {
				h.log.Debug(fmt.Sprintf(
					"checkAndReSendWithProposalContent QueryTxByHash failed. will query after %f second",
					BlockRetryInterval.Seconds()),
					"tx hash", txHashStr,
					"err", err)
			}

			time.Sleep(BlockRetryInterval)
			retry--
			continue
		}

		if res.Code != 0 {
			switch {
			case strings.Contains(res.RawLog, stafiHubXRelayersTypes.ErrAlreadyVoted.Error()):
				h.log.Info("no need send, already voted", "txHash", txHashStr, "type", typeStr)
				return nil
			case strings.Contains(res.RawLog, stafiHubXRVoteTypes.ErrProposalAlreadyApproved.Error()):
				h.log.Info("no need send, already approved", "txHash", txHashStr, "type", typeStr)
				return nil
			case strings.Contains(res.RawLog, stafiHubXRVoteTypes.ErrProposalAlreadyExpired.Error()):
				h.log.Info("no need send, already expired", "txHash", txHashStr, "type", typeStr)
				return nil
			case strings.Contains(res.RawLog, stafiHubXLedgerTypes.ErrEraNotContinuable.Error()):
				h.log.Info("no need send, already update new era", "txHash", txHashStr, "type", typeStr)
				return nil

			// resend case
			case strings.Contains(res.RawLog, errors.ErrOutOfGas.Error()):
				return h.checkAndReSendWithProposalContent(txHashStr, content)
			default:
				return fmt.Errorf("tx failed, txHash: %s, rawlog: %s", txHashStr, res.RawLog)
			}
		}

		break
	}

	h.log.Info("checkAndReSendWithProposalContent success", "txHash", txHashStr, "type", typeStr)
	return nil
}
