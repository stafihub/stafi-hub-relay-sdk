package client

import (
	"fmt"

	"github.com/stafiprotocol/rtoken-relay-core/common/core"
	stafiHubXRvoteTypes "github.com/stafiprotocol/stafihub/x/rvote/types"
)

func (c *Client) SubmitProposal(content stafiHubXRvoteTypes.Content) (string, error) {
	done := core.UseSdkConfigContext(AccountPrefix)
	msg, err := stafiHubXRvoteTypes.NewMsgSubmitProposal(c.GetFromAddress(), content)
	if err != nil {
		done()
		return "", fmt.Errorf("stafiHubXRvoteTypes.NewMsgSubmitProposal faild: %s", err)
	}

	if err := msg.ValidateBasic(); err != nil {
		done()
		return "", fmt.Errorf("msg.ValidateBasic faild: %s", err)
	}
	done()

	txBts, err := c.ConstructAndSignTx(msg)
	if err != nil {
		return "", fmt.Errorf("c.ConstructAndSignTx faild: %s", err)
	}
	return c.BroadcastTx(txBts)
}
