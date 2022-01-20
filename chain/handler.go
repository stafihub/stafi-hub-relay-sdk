package chain

import (
	"fmt"

	"github.com/ChainSafe/log15"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/stafiprotocol/rtoken-relay-core/core"
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
	default:
		return fmt.Errorf("message reason unsupported reason: %s", m.Reason)
	}
	return nil
}

func (w *Handler) handleExeLiquidityBond(m *core.Message) error {
	proposal, ok := m.Content.(core.ProposalExeLiquidityBond)
	if !ok {
		return fmt.Errorf("ProposalLiquidityBond cast failed, %+v", m)
	}
	bonder, err := types.AccAddressFromBech32(proposal.Bonder)
	if err != nil {
		return err
	}
	content := stafiHubXLedgerTypes.NewExecuteBondProposal(w.conn.client.GetFromAddress(), proposal.Denom, bonder, proposal.Pool, proposal.Blockhash, proposal.Txhash, proposal.Amount)
	txHash, err := w.conn.SubmitProposal(content)
	if err != nil {
		return err
	}
	w.log.Info("submitProposl", "tx hash", txHash)
	return nil
}

func (w *Handler) handleNewChainEra(m *core.Message) error {
	proposal, ok := m.Content.(core.ProposalSetChainEra)
	if !ok {
		return fmt.Errorf("ProposalSetChainEra cast failed, %+v", m)
	}
	content := stafiHubXLedgerTypes.NewSetChainEraProposal(w.conn.client.GetFromAddress(), proposal.Denom, proposal.Era)
	txHash, err := w.conn.SubmitProposal(content)
	if err != nil {
		return err
	}
	w.log.Info("submitProposl", "tx hash", txHash)
	return nil
}

func (w *Handler) handleBondReport(m *core.Message) error {
	proposal, ok := m.Content.(core.ProposalBondReport)
	if !ok {
		return fmt.Errorf("ProposalBondReport cast failed, %+v", m)
	}
	content := stafiHubXLedgerTypes.NewBondReportProposal(w.conn.client.GetFromAddress(), proposal.Denom, proposal.ShotId, proposal.Action)
	txHash, err := w.conn.SubmitProposal(content)
	if err != nil {
		return err
	}
	w.log.Info("submitProposl", "tx hash", txHash)
	return nil
}

func (w *Handler) handleActiveReport(m *core.Message) error {
	proposal, ok := m.Content.(core.ProposalActiveReport)
	if !ok {
		return fmt.Errorf("ProposalActiveReport cast failed, %+v", m)
	}
	content := stafiHubXLedgerTypes.NewActiveReportProposal(w.conn.client.GetFromAddress(), proposal.Denom, proposal.ShotId, proposal.Staked, proposal.Unstaked)
	txHash, err := w.conn.SubmitProposal(content)
	if err != nil {
		return err
	}
	w.log.Info("submitProposl", "tx hash", txHash)
	return nil
}

func (w *Handler) handleWithdrawReport(m *core.Message) error {
	proposal, ok := m.Content.(core.ProposalWithdrawReport)
	if !ok {
		return fmt.Errorf("ProposalWithdrawReport cast failed, %+v", m)
	}
	content := stafiHubXLedgerTypes.NewWithdrawReportProposal(w.conn.client.GetFromAddress(), proposal.Denom, proposal.ShotId)
	txHash, err := w.conn.SubmitProposal(content)
	if err != nil {
		return err
	}
	w.log.Info("submitProposl", "tx hash", txHash)
	return nil
}

func (w *Handler) handleTransferReport(m *core.Message) error {
	proposal, ok := m.Content.(core.ProposalTransferReport)
	if !ok {
		return fmt.Errorf("ProposalTransferReport cast failed, %+v", m)
	}
	content := stafiHubXLedgerTypes.NewTransferReportProposal(w.conn.client.GetFromAddress(), proposal.Denom, proposal.ShotId)
	txHash, err := w.conn.SubmitProposal(content)
	if err != nil {
		return err
	}
	w.log.Info("submitProposl", "tx hash", txHash)
	return nil
}
