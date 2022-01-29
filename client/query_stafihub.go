package client

import (
	"context"

	"github.com/stafiprotocol/rtoken-relay-core/common/core"
	stafiHubXLedgerTypes "github.com/stafiprotocol/stafihub/x/ledger/types"
)

func (c *Client) QuerySnapshot(shotId []byte) (*stafiHubXLedgerTypes.QueryGetSnapshotResponse, error) {
	done := core.UseSdkConfigContext(AccountPrefix)
	defer done()

	queryClient := stafiHubXLedgerTypes.NewQueryClient(c.Ctx())
	params := &stafiHubXLedgerTypes.QueryGetSnapshotRequest{
		ShotId: shotId,
	}

	cc, err := Retry(func() (interface{}, error) {
		return queryClient.GetSnapshot(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXLedgerTypes.QueryGetSnapshotResponse), nil
}

func (c *Client) QueryPoolUnbond(denom, pool string, era uint32) (*stafiHubXLedgerTypes.QueryGetPoolUnbondResponse, error) {
	done := core.UseSdkConfigContext(AccountPrefix)
	defer done()

	queryClient := stafiHubXLedgerTypes.NewQueryClient(c.Ctx())
	params := &stafiHubXLedgerTypes.QueryGetPoolUnbondRequest{
		Denom: denom,
		Pool:  pool,
		Era:   era,
	}

	cc, err := Retry(func() (interface{}, error) {
		return queryClient.GetPoolUnbond(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXLedgerTypes.QueryGetPoolUnbondResponse), nil
}

func (c *Client) QueryPoolDetail(denom, pool string) (*stafiHubXLedgerTypes.QueryGetPoolDetailResponse, error) {
	done := core.UseSdkConfigContext(AccountPrefix)
	defer done()

	queryClient := stafiHubXLedgerTypes.NewQueryClient(c.Ctx())
	params := &stafiHubXLedgerTypes.QueryGetPoolDetailRequest{
		Denom: denom,
		Pool:  pool,
	}

	cc, err := Retry(func() (interface{}, error) {
		return queryClient.GetPoolDetail(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXLedgerTypes.QueryGetPoolDetailResponse), nil
}

func (c *Client) QuerySignature(denom, pool string, era uint32, txType stafiHubXLedgerTypes.OriginalTxType, proposalId []byte) (*stafiHubXLedgerTypes.QueryGetSignatureResponse, error) {
	done := core.UseSdkConfigContext(AccountPrefix)
	defer done()

	queryClient := stafiHubXLedgerTypes.NewQueryClient(c.Ctx())
	params := &stafiHubXLedgerTypes.QueryGetSignatureRequest{
		Denom:  denom,
		Era:    era,
		Pool:   pool,
		TxType: txType,
		PropId: proposalId,
	}

	cc, err := Retry(func() (interface{}, error) {
		return queryClient.GetSignature(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXLedgerTypes.QueryGetSignatureResponse), nil
}

func (c *Client) QueryPools(denom string) (*stafiHubXLedgerTypes.QueryPoolsByDenomResponse, error) {
	done := core.UseSdkConfigContext(AccountPrefix)
	defer done()

	queryClient := stafiHubXLedgerTypes.NewQueryClient(c.Ctx())
	params := &stafiHubXLedgerTypes.QueryPoolsByDenomRequest{
		Denom: denom,
	}

	cc, err := Retry(func() (interface{}, error) {
		return queryClient.PoolsByDenom(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXLedgerTypes.QueryPoolsByDenomResponse), nil
}
