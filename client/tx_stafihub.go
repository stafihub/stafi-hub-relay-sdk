package client

import (
	"fmt"

	clientTx "github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stafihub/rtoken-relay-core/common/core"
	stafiHubXLedgerTypes "github.com/stafihub/stafihub/x/ledger/types"
	stafiHubXRvoteTypes "github.com/stafihub/stafihub/x/rvote/types"
)

func (c *Client) SubmitProposal(content stafiHubXRvoteTypes.Content) (string, []byte, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	msg, err := stafiHubXRvoteTypes.NewMsgSubmitProposal(c.GetFromAddress(), content)
	if err != nil {
		done()
		return "", nil, fmt.Errorf("stafiHubXRvoteTypes.NewMsgSubmitProposal faild: %s", err)
	}

	if err := msg.ValidateBasic(); err != nil {
		done()
		return "", nil, fmt.Errorf("msg.ValidateBasic faild: %s", err)
	}
	done()
	txBts, err := c.ConstructAndSignTx(msg)
	if err != nil {
		return "", nil, fmt.Errorf("c.ConstructAndSignTx faild: %s", err)
	}
	txHash, err := c.BroadcastTx(txBts)
	return txHash, txBts, err
}

func (c *Client) SubmitSignature(sigMsg *stafiHubXLedgerTypes.MsgSubmitSignature) (string, []byte, error) {
	txBts, err := c.ConstructAndSignTx(sigMsg)
	if err != nil {
		return "", nil, fmt.Errorf("c.ConstructAndSignTx faild: %s", err)
	}
	txHash, err := c.BroadcastTx(txBts)
	return txHash, txBts, err
}

func (c *Client) CalculateGas(txf clientTx.Factory, msgs ...sdk.Msg) (uint64, error) {

	cc, err := c.retry(func() (interface{}, error) {
		_, adjustGas, err := clientTx.CalculateGas(c.Ctx(), txf, msgs...)
		return adjustGas, err
	})
	if err != nil {
		return 0, err
	}

	return cc.(uint64), err
}
