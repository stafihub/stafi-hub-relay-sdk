package client

import (
	"fmt"

	clientTx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xAuthClient "github.com/cosmos/cosmos-sdk/x/auth/client"
	xBankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/spf13/cobra"
	"github.com/stafihub/rtoken-relay-core/common/core"
)

func (c *Client) SingleTransferTo(toAddr types.AccAddress, amount types.Coins) (string, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	msg := xBankTypes.NewMsgSend(c.Ctx().GetFromAddress(), toAddr, amount)
	done()

	txBts, err := c.ConstructAndSignTx(msg)
	if err != nil {
		return "", err
	}

	return c.BroadcastTx(txBts)
}

func (c *Client) BroadcastBatchMsg(msgs []types.Msg) (string, error) {
	txBts, err := c.ConstructAndSignTx(msgs...)
	if err != nil {
		return "", err
	}
	return c.BroadcastTx(txBts)
}

func (c *Client) BroadcastTx(tx []byte) (string, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.retry(func() (interface{}, error) {
		return c.Ctx().BroadcastTx(tx)
	})
	if err != nil {
		return "", fmt.Errorf("retry broadcastTx err: %s", err)
	}
	res := cc.(*types.TxResponse)
	if res.Code != 0 {
		return "", fmt.Errorf("broadcast err, res.codespace: %s, res.code: %d, res.raw_log: %s", res.Codespace, res.Code, res.RawLog)
	}
	return res.TxHash, nil
}

func (c *Client) ConstructAndSignTx(msgs ...types.Msg) ([]byte, error) {
	account, err := c.GetAccount()
	if err != nil {
		return nil, err
	}
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	clientCtx := c.Ctx()

	cmd := cobra.Command{}
	txf, err := clientTx.NewFactoryCLI(clientCtx, cmd.Flags())
	if err != nil {
		return nil, err
	}
	txf = txf.WithSequence(account.GetSequence()).
		WithAccountNumber(account.GetAccountNumber()).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT). //multi sig need this mod
		WithGasAdjustment(1.5).
		WithGas(0).
		WithGasPrices(c.gasPrice).
		WithSimulateAndExecute(true)

	// auto cal gas with retry
	adjusted, err := c.CalculateGas(txf, msgs...)
	if err != nil {
		return nil, fmt.Errorf("client.CalculateGas failed: %s", err)
	}
	txf = txf.WithGas(adjusted * 2)

	txBuilderRaw, err := txf.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, fmt.Errorf("clientTx.BuildUnsignedTx faild: %s", err)
	}

	err = xAuthClient.SignTx(txf, clientCtx, clientCtx.GetFromName(), txBuilderRaw, true, true)
	if err != nil {
		return nil, fmt.Errorf("xAuthClient.SignTx failed: %s", err)
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(txBuilderRaw.GetTx())
	if err != nil {
		return nil, fmt.Errorf("TxConfig.TxEncoder failed: %s", err)
	}
	return txBytes, nil
}
