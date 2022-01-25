package client

import (
	"context"
	stafiHubXLedgerTypes "github.com/stafiprotocol/stafihub/x/ledger/types"
)

func (c *Client) QuerySnapshot(shotId []byte) (*stafiHubXLedgerTypes.QueryGetSnapshotResponse, error) {
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
