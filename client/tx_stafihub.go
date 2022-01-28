package client

import (
	"github.com/stafiprotocol/rtoken-relay-core/common/core"
	stafiHubXRvoteTypes "github.com/stafiprotocol/stafihub/x/rvote/types"
)

func (c *Client) SubmitProposal(content stafiHubXRvoteTypes.Content) (string, error) {
	done := core.UseSdkConfigContext(AccountPrefix)
	msg, err := stafiHubXRvoteTypes.NewMsgSubmitProposal(c.GetFromAddress(), content)
	if err != nil {
		done()
		return "", err
	}

	if err := msg.ValidateBasic(); err != nil {
		done()
		return "", err
	}
	done()

	txBts, err := c.ConstructAndSignTx(msg)
	if err != nil {
		return "", err
	}
	return c.BroadcastTx(txBts)
}
