package chain

import (
	"fmt"

	"github.com/ChainSafe/log15"
	"github.com/stafiprotocol/rtoken-relay-core/core"
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
			ok := w.handleMessage(msg)
			if !ok {
				w.sysErrChan <- fmt.Errorf("resolveMessage process failed. %+v", msg)
				return
			}
		}
	}
}

//resolve msg from other chains
func (w *Handler) handleMessage(m *core.Message) (processOk bool) {
	switch m.Reason {
	case core.ReasonLiquidityBondResult:
	case core.ReasonBondReportedEvent:
	case core.ReasonActiveReport:
	case core.ReasonWithdrawReport:
	case core.ReasonTransferReport:
	case core.ReasonSubmitSignature:
	default:
		w.log.Warn("message reason unsupported", "reason", m.Reason)
		return false
	}
	return true
}
