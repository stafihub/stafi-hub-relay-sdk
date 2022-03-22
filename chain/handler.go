package chain

import (
	"fmt"
	"strings"
	"time"

	"github.com/ChainSafe/log15"
	"github.com/cosmos/cosmos-sdk/types"
	errType "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stafihub/rtoken-relay-core/common/core"
	hubClient "github.com/stafihub/stafi-hub-relay-sdk/client"
	stafiHubXLedgerTypes "github.com/stafihub/stafihub/x/ledger/types"
	stafiHubXRelayersTypes "github.com/stafihub/stafihub/x/relayers/types"
	stafiHubXRVoteTypes "github.com/stafihub/stafihub/x/rvote/types"
	"google.golang.org/grpc/codes"
)

const msgLimit = 4096

type Handler struct {
	conn       *Connection
	router     *core.Router
	msgChan    chan *core.Message
	log        log15.Logger
	stopChan   <-chan struct{}
	sysErrChan chan<- error
}

func NewHandler(conn *Connection, log log15.Logger, stopChan <-chan struct{}, sysErrChan chan<- error) *Handler {
	return &Handler{
		conn:       conn,
		msgChan:    make(chan *core.Message, msgLimit),
		log:        log,
		stopChan:   stopChan,
		sysErrChan: sysErrChan,
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
	}
	// deal write msg
	w.queueMessage(m)
}

func (w *Handler) queueMessage(m *core.Message) {
	w.msgChan <- m
}

func (w *Handler) msgHandler() {
	for {
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
		}
	}
}

//resolve write msg from other chains
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
	case core.ReasonWithdrawReport:
		return w.handleWithdrawReport(m)
	case core.ReasonTransferReport:
		return w.handleTransferReport(m)
	case core.ReasonSubmitSignature:
		return w.handleSubmitSignature(m)
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
	_, err := w.conn.client.QueryBondRecord(proposal.Denom, proposal.Txhash)
	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		return err
	}
	if err == nil {
		w.log.Warn("handleExeLiquidityBond already exe bond, no need submitproposal", "msg", m)
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

	txHash, txBts, err := w.conn.client.SubmitProposal(content)
	return w.checkAndReSend(txHash, txBts, "executeBondProposal", err)
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

	txHash, txBts, err := w.conn.client.SubmitProposal(content)
	return w.checkAndReSend(txHash, txBts, "setChainEraProposal", err)
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

	txHash, txBts, err := w.conn.client.SubmitProposal(content)
	return w.checkAndReSend(txHash, txBts, "bondReportProposal", err)
}

func (w *Handler) handleActiveReport(m *core.Message) error {
	w.log.Info("handleActiveReport", "m", m)
	proposal, ok := m.Content.(core.ProposalActiveReport)
	if !ok {
		return fmt.Errorf("ProposalActiveReport cast failed, %+v", m)
	}

	done := core.UseSdkConfigContext(hubClient.GetAccountPrefix())
	content := stafiHubXLedgerTypes.NewActiveReportProposal(w.conn.client.GetFromAddress(), proposal.Denom, proposal.ShotId, proposal.Staked, proposal.Unstaked)
	done()

	txHash, txBts, err := w.conn.client.SubmitProposal(content)
	return w.checkAndReSend(txHash, txBts, "activeReportProposal", err)
}

func (w *Handler) handleWithdrawReport(m *core.Message) error {
	w.log.Info("handleWithdrawReport", "m", m)
	proposal, ok := m.Content.(core.ProposalWithdrawReport)
	if !ok {
		return fmt.Errorf("ProposalWithdrawReport cast failed, %+v", m)
	}

	done := core.UseSdkConfigContext(hubClient.GetAccountPrefix())
	content := stafiHubXLedgerTypes.NewWithdrawReportProposal(w.conn.client.GetFromAddress(), proposal.Denom, proposal.ShotId)
	done()

	txHash, txBts, err := w.conn.client.SubmitProposal(content)
	return w.checkAndReSend(txHash, txBts, "withdrawReportProposal", err)
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

	txHash, txBts, err := w.conn.client.SubmitProposal(content)
	return w.checkAndReSend(txHash, txBts, "transferReportProposal", err)
}

func (w *Handler) handleSubmitSignature(m *core.Message) error {
	w.log.Info("handleSubmitSignature", "m", m)
	proposal, ok := m.Content.(core.ParamSubmitSignature)
	if !ok {
		return fmt.Errorf("ProposalSubmitSignature cast failed, %+v", m)
	}

	done := core.UseSdkConfigContext(hubClient.GetAccountPrefix())
	msg := stafiHubXLedgerTypes.NewMsgSubmitSignature(w.conn.client.GetFromAddress().String(), proposal.Denom, proposal.Era, proposal.Pool, proposal.TxType, proposal.PropId, proposal.Signature)
	done()

	txHash, txBts, err := w.conn.client.SubmitSignature(msg)
	return w.checkAndReSend(txHash, txBts, "submitSignature", err)
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
	w.log.Info("handleGetSignature", "m", m)
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

	w.log.Info("getSignatures", "sigs", sigs.Signature.Sigs)
	return nil
}

func (h *Handler) checkAndReSend(txHashStr string, txBts []byte, typeStr string, err error) error {
	if err != nil {
		switch {
		case strings.Contains(err.Error(), stafiHubXLedgerTypes.ErrSignatureRepeated.Error()):
			h.log.Info("no need send, already submit signature", "txHash", txHashStr, "type", typeStr)
			return nil
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
		}

		return err
	}

	retry := BlockRetryLimit
	for {
		if retry <= 0 {
			return fmt.Errorf("checkAndSend broadcast tx reach retry limit, tx hash: %s", txHashStr)
		}
		//check on chain
		res, err := h.conn.client.QueryTxByHash(txHashStr)
		if err != nil || res.Empty() || res.Code != 0 {
			if res != nil {
				h.log.Warn(fmt.Sprintf(
					"checkAndSend QueryTxByHash, tx failed. will rebroadcast after %f second",
					BlockRetryInterval.Seconds()),
					"tx hash", txHashStr,
					"res.code", res.Code)
			} else {
				h.log.Warn(fmt.Sprintf(
					"checkAndSend QueryTxByHash failed. will rebroadcast after %f second",
					BlockRetryInterval.Seconds()),
					"tx hash", txHashStr,
					"err", err)
			}

			//broadcast if not on chain
			_, err = h.conn.client.BroadcastTx(txBts)
			if err != nil && err != errType.ErrTxInMempoolCache {
				h.log.Warn("checkAndSend BroadcastTx failed  will retry", "failed info", err)
			}
			time.Sleep(BlockRetryInterval)
			retry--
			continue
		}
		break
	}

	h.log.Info("checkAndSend success", "txHash", txHashStr, "type", typeStr)
	return nil
}
