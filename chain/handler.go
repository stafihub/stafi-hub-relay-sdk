package chain

import (
	"fmt"
	"strings"
	"time"

	"github.com/ChainSafe/log15"
	"github.com/cosmos/cosmos-sdk/types"
	errType "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stafiprotocol/rtoken-relay-core/common/core"
	hubClient "github.com/stafiprotocol/stafi-hub-relay-sdk/client"
	stafiHubXLedgerTypes "github.com/stafiprotocol/stafihub/x/ledger/types"
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

//resolve msg from other chains
func (w *Handler) HandleMessage(m *core.Message) {
	switch m.Reason {
	case core.ReasonGetPools:
		go w.handleGetPools(m)
		return
	}
	w.queueMessage(m)
}

func (w *Handler) queueMessage(m *core.Message) {
	w.msgChan <- m
}

func (w *Handler) msgHandler() {
	for {
		select {
		case <-w.stopChan:
			w.log.Info("msgHandler stop")
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

//resolve msg from other chains
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
	w.log.Info("handleExeLiquidityBond", "msg", m)
	proposal, ok := m.Content.(core.ProposalExeLiquidityBond)
	if !ok {
		return fmt.Errorf("ProposalLiquidityBond cast failed, %+v", m)
	}

	done := core.UseSdkConfigContext(hubClient.AccountPrefix)
	bonder, err := types.AccAddressFromBech32(proposal.Bonder)
	if err != nil {
		done()
		return err
	}
	content := stafiHubXLedgerTypes.NewExecuteBondProposal(w.conn.client.GetFromAddress(), proposal.Denom, bonder, proposal.Pool, proposal.Blockhash, proposal.Txhash, proposal.Amount)
	done()

	txHash, txBts, err := w.conn.client.SubmitProposal(content)
	return w.checkAndReSend(txHash, txBts, "executeBondProposal", err)
}

func (w *Handler) handleNewChainEra(m *core.Message) error {
	w.log.Debug("handleNewChainEra", "msg", m)
	proposal, ok := m.Content.(core.ProposalSetChainEra)
	if !ok {
		return fmt.Errorf("ProposalSetChainEra cast failed, %+v", m)
	}

	chainEra, err := w.conn.client.QueryChainEra(proposal.Denom)
	if err != nil {
		return err
	}

	if chainEra.GetEra() >= proposal.Era {
		return nil
	}
	continuable, err := w.conn.client.QueryEraContinuable(proposal.Denom, chainEra.GetEra())
	if err != nil {
		return err
	}
	if !continuable {
		return nil
	}

	useEra := chainEra.GetEra() + 1
	w.log.Info("will set newChainEra", "new era", useEra, "symbol", proposal.Denom)

	done := core.UseSdkConfigContext(hubClient.AccountPrefix)
	content := stafiHubXLedgerTypes.NewSetChainEraProposal(w.conn.client.GetFromAddress(), proposal.Denom, useEra)
	done()

	txHash, txBts, err := w.conn.client.SubmitProposal(content)
	return w.checkAndReSend(txHash, txBts, "setChainEraProposal", err)
}

func (w *Handler) handleBondReport(m *core.Message) error {
	w.log.Info("handleBondReport", "msg", m)
	proposal, ok := m.Content.(core.ProposalBondReport)
	if !ok {
		return fmt.Errorf("ProposalBondReport cast failed, %+v", m)
	}

	done := core.UseSdkConfigContext(hubClient.AccountPrefix)
	content := stafiHubXLedgerTypes.NewBondReportProposal(w.conn.client.GetFromAddress(), proposal.Denom, proposal.ShotId, proposal.Action)
	done()

	txHash, txBts, err := w.conn.client.SubmitProposal(content)
	return w.checkAndReSend(txHash, txBts, "bondReportProposal", err)
}

func (w *Handler) handleActiveReport(m *core.Message) error {
	w.log.Info("handleActiveReport", "msg", m)
	proposal, ok := m.Content.(core.ProposalActiveReport)
	if !ok {
		return fmt.Errorf("ProposalActiveReport cast failed, %+v", m)
	}

	done := core.UseSdkConfigContext(hubClient.AccountPrefix)
	content := stafiHubXLedgerTypes.NewActiveReportProposal(w.conn.client.GetFromAddress(), proposal.Denom, proposal.ShotId, proposal.Staked, proposal.Unstaked)
	done()

	txHash, txBts, err := w.conn.client.SubmitProposal(content)
	return w.checkAndReSend(txHash, txBts, "activeReportProposal", err)
}

func (w *Handler) handleWithdrawReport(m *core.Message) error {
	w.log.Info("handleWithdrawReport", "msg", m)
	proposal, ok := m.Content.(core.ProposalWithdrawReport)
	if !ok {
		return fmt.Errorf("ProposalWithdrawReport cast failed, %+v", m)
	}

	done := core.UseSdkConfigContext(hubClient.AccountPrefix)
	content := stafiHubXLedgerTypes.NewWithdrawReportProposal(w.conn.client.GetFromAddress(), proposal.Denom, proposal.ShotId)
	done()

	txHash, txBts, err := w.conn.client.SubmitProposal(content)
	return w.checkAndReSend(txHash, txBts, "withdrawReportProposal", err)
}

func (w *Handler) handleTransferReport(m *core.Message) error {
	w.log.Info("handleTransferReport", "msg", m)
	proposal, ok := m.Content.(core.ProposalTransferReport)
	if !ok {
		return fmt.Errorf("ProposalTransferReport cast failed, %+v", m)
	}

	done := core.UseSdkConfigContext(hubClient.AccountPrefix)
	content := stafiHubXLedgerTypes.NewTransferReportProposal(w.conn.client.GetFromAddress(), proposal.Denom, proposal.ShotId)
	done()

	txHash, txBts, err := w.conn.client.SubmitProposal(content)
	return w.checkAndReSend(txHash, txBts, "transferReportProposal", err)
}

func (w *Handler) handleSubmitSignature(m *core.Message) error {
	w.log.Info("handleSubmitSignature", "msg", m)
	proposal, ok := m.Content.(core.ParamSubmitSignature)
	if !ok {
		return fmt.Errorf("ProposalSubmitSignature cast failed, %+v", m)
	}

	done := core.UseSdkConfigContext(hubClient.AccountPrefix)
	msg := stafiHubXLedgerTypes.NewMsgSubmitSignature(w.conn.client.GetFromAddress().String(), proposal.Denom, proposal.Era, proposal.Pool, proposal.TxType, proposal.PropId, proposal.Signature)
	done()

	txHash, txBts, err := w.conn.client.SubmitSignature(msg)
	return w.checkAndReSend(txHash, txBts, "submitSignature", err)
}

func (w *Handler) handleGetPools(m *core.Message) error {
	w.log.Info("handleGetPools", "msg", m)
	getPools, ok := m.Content.(core.ParamGetPools)
	if !ok {
		return fmt.Errorf("GetPools cast failed, %+v", m)
	}
	pools, err := w.conn.client.QueryPools(getPools.Denom)
	if err != nil {
		getPools.Pools <- []string{}
		return err
	}
	getPools.Pools <- pools.GetAddrs()

	w.log.Info("getPools", "pools", pools.GetAddrs())
	return nil
}

func (h *Handler) checkAndReSend(txHashStr string, txBts []byte, typeStr string, err error) error {
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "signature repeated"):
			h.log.Info("no need send, already submit signature", "txHash", txHashStr, "type", typeStr)
			return nil
		}
		return err
	} else {
		retry := BlockRetryLimit
		for {
			if retry <= 0 {
				return fmt.Errorf("checkAndSend broadcast tx reach retry limit, tx hash: %s", txHashStr)
			}
			//check on chain
			res, err := h.conn.client.QueryTxByHash(txHashStr)
			if err != nil || res.Empty() || res.Code != 0 {
				h.log.Warn(fmt.Sprintf(
					"checkAndSend QueryTxByHash failed. will rebroadcast after %f second",
					BlockRetryInterval.Seconds()),
					"tx hash", txHashStr,
					"err or res.empty", err)

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
	}
	h.log.Info("checkAndSend success", "txHash", txHashStr, "type", typeStr)
	return nil
}
