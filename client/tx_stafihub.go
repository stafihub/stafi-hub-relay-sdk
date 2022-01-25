package client

import (
	stafiHubXRvoteTypes "github.com/stafiprotocol/stafihub/x/rvote/types"
)

func (c *Client) SubmitProposal(content stafiHubXRvoteTypes.Content) (string, error) {
	msg, err := stafiHubXRvoteTypes.NewMsgSubmitProposal(c.GetFromAddress(), content)
	if err != nil {
		return "", err
	}

	if err := msg.ValidateBasic(); err != nil {
		return "", err
	}
	txBts, err := c.ConstructAndSignTx(msg)
	if err != nil {
		return "", err
	}
	return c.BroadcastTx(txBts)
}
