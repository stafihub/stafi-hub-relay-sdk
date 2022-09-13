package client

import (
	"context"

	"github.com/stafihub/rtoken-relay-core/common/core"
	stafiHubXBridgeTypes "github.com/stafihub/stafihub/x/bridge/types"
	stafiHubXLedgerTypes "github.com/stafihub/stafihub/x/ledger/types"
	stafiHubXRBankTypes "github.com/stafihub/stafihub/x/rbank/types"
	stafiHubXRValidatorTypes "github.com/stafihub/stafihub/x/rvalidator/types"
)

func (c *Client) QuerySnapshot(shotId string) (*stafiHubXLedgerTypes.QueryGetSnapshotResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.Retry(func() (interface{}, error) {
		queryClient := stafiHubXLedgerTypes.NewQueryClient(c.Ctx())
		params := &stafiHubXLedgerTypes.QueryGetSnapshotRequest{
			ShotId: shotId,
		}
		return queryClient.GetSnapshot(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXLedgerTypes.QueryGetSnapshotResponse), nil
}

func (c *Client) QueryPoolUnbond(denom, pool string, era uint32) (*stafiHubXLedgerTypes.QueryPoolUnbondingsResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.Retry(func() (interface{}, error) {
		queryClient := stafiHubXLedgerTypes.NewQueryClient(c.Ctx())
		params := &stafiHubXLedgerTypes.QueryPoolUnbondingsRequest{
			Denom:     denom,
			Pool:      pool,
			UnlockEra: era,
		}
		return queryClient.PoolUnbondings(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXLedgerTypes.QueryPoolUnbondingsResponse), nil
}

func (c *Client) QueryPoolDetail(denom, pool string) (*stafiHubXLedgerTypes.QueryGetPoolDetailResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.Retry(func() (interface{}, error) {
		queryClient := stafiHubXLedgerTypes.NewQueryClient(c.Ctx())
		params := &stafiHubXLedgerTypes.QueryGetPoolDetailRequest{
			Denom: denom,
			Pool:  pool,
		}
		return queryClient.GetPoolDetail(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXLedgerTypes.QueryGetPoolDetailResponse), nil
}

func (c *Client) QuerySignature(denom, pool string, era uint32, txType stafiHubXLedgerTypes.OriginalTxType, proposalId string) (*stafiHubXLedgerTypes.QueryGetSignatureResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.Retry(func() (interface{}, error) {
		queryClient := stafiHubXLedgerTypes.NewQueryClient(c.Ctx())
		params := &stafiHubXLedgerTypes.QueryGetSignatureRequest{
			Denom:  denom,
			Era:    era,
			Pool:   pool,
			TxType: txType,
			PropId: proposalId,
		}
		return queryClient.GetSignature(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXLedgerTypes.QueryGetSignatureResponse), nil
}

func (c *Client) QueryPools(denom string) (*stafiHubXLedgerTypes.QueryBondedPoolsByDenomResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.Retry(func() (interface{}, error) {
		queryClient := stafiHubXLedgerTypes.NewQueryClient(c.Ctx())
		params := &stafiHubXLedgerTypes.QueryBondedPoolsByDenomRequest{
			Denom: denom,
		}
		return queryClient.BondedPoolsByDenom(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXLedgerTypes.QueryBondedPoolsByDenomResponse), nil
}

func (c *Client) QueryChainEra(denom string) (*stafiHubXLedgerTypes.QueryGetChainEraResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.Retry(func() (interface{}, error) {
		queryClient := stafiHubXLedgerTypes.NewQueryClient(c.Ctx())
		params := &stafiHubXLedgerTypes.QueryGetChainEraRequest{
			Denom: denom,
		}
		return queryClient.GetChainEra(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXLedgerTypes.QueryGetChainEraResponse), nil
}

func (c *Client) QueryEraSnapShotList(denom string, era uint32) (*stafiHubXLedgerTypes.QueryGetEraSnapshotResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.Retry(func() (interface{}, error) {
		queryClient := stafiHubXLedgerTypes.NewQueryClient(c.Ctx())
		params := &stafiHubXLedgerTypes.QueryGetEraSnapshotRequest{
			Denom: denom,
			Era:   era,
		}
		return queryClient.GetEraSnapshot(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXLedgerTypes.QueryGetEraSnapshotResponse), nil
}

func (c *Client) QueryEraContinuable(denom string, era uint32) (bool, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.Retry(func() (interface{}, error) {
		queryClient := stafiHubXLedgerTypes.NewQueryClient(c.Ctx())
		params := &stafiHubXLedgerTypes.QueryGetEraSnapshotRequest{
			Denom: denom,
			Era:   era,
		}
		return queryClient.GetEraSnapshot(context.Background(), params)
	})
	if err != nil {
		return false, err
	}
	res := cc.(*stafiHubXLedgerTypes.QueryGetEraSnapshotResponse)
	if len(res.ShotIds) > 0 {
		return false, nil
	}
	return true, nil
}

func (c *Client) QueryEraRate(denom string, era uint32) (*stafiHubXLedgerTypes.QueryGetEraExchangeRateResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.Retry(func() (interface{}, error) {
		queryClient := stafiHubXLedgerTypes.NewQueryClient(c.Ctx())
		params := &stafiHubXLedgerTypes.QueryGetEraExchangeRateRequest{
			Denom: denom,
			Era:   era,
		}
		return queryClient.GetEraExchangeRate(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXLedgerTypes.QueryGetEraExchangeRateResponse), nil
}

func (c *Client) QueryRate(denom string) (*stafiHubXLedgerTypes.QueryGetExchangeRateResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.Retry(func() (interface{}, error) {
		queryClient := stafiHubXLedgerTypes.NewQueryClient(c.Ctx())
		params := &stafiHubXLedgerTypes.QueryGetExchangeRateRequest{
			Denom: denom,
		}
		return queryClient.GetExchangeRate(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXLedgerTypes.QueryGetExchangeRateResponse), nil
}

func (c *Client) QueryRParams(denom string) (*stafiHubXLedgerTypes.QueryGetRParamsResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.Retry(func() (interface{}, error) {
		queryClient := stafiHubXLedgerTypes.NewQueryClient(c.Ctx())
		params := &stafiHubXLedgerTypes.QueryGetRParamsRequest{
			Denom: denom,
		}
		return queryClient.GetRParams(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXLedgerTypes.QueryGetRParamsResponse), nil
}

func (c *Client) QueryBondRecord(denom, txHash string) (*stafiHubXLedgerTypes.QueryGetBondRecordResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.Retry(func() (interface{}, error) {
		queryClient := stafiHubXLedgerTypes.NewQueryClient(c.Ctx())
		params := &stafiHubXLedgerTypes.QueryGetBondRecordRequest{
			Denom:  denom,
			Txhash: txHash,
		}
		return queryClient.GetBondRecord(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXLedgerTypes.QueryGetBondRecordResponse), nil
}

func (c *Client) QueryAddressPrefix(denom string) (*stafiHubXRBankTypes.QueryAddressPrefixResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.Retry(func() (interface{}, error) {
		queryClient := stafiHubXRBankTypes.NewQueryClient(c.Ctx())
		params := &stafiHubXRBankTypes.QueryAddressPrefixRequest{
			Denom: denom,
		}
		return queryClient.AddressPrefix(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXRBankTypes.QueryAddressPrefixResponse), nil
}

func (c *Client) QueryBridgeProposalDetail(chainId uint32, depositNonce uint64, resourceId, amount, receiver string) (*stafiHubXBridgeTypes.QueryProposalDetailResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.Retry(func() (interface{}, error) {
		queryClient := stafiHubXBridgeTypes.NewQueryClient(c.Ctx())
		params := &stafiHubXBridgeTypes.QueryProposalDetailRequest{
			ChainId:      chainId,
			DepositNonce: depositNonce,
			ResourceId:   resourceId,
			Amount:       amount,
			Receiver:     receiver,
		}
		return queryClient.ProposalDetail(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXBridgeTypes.QueryProposalDetailResponse), nil
}

func (c *Client) QueryCycleSeconds(denom string) (*stafiHubXRValidatorTypes.QueryCycleSecondsResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.Retry(func() (interface{}, error) {
		queryClient := stafiHubXRValidatorTypes.NewQueryClient(c.Ctx())
		return queryClient.CycleSeconds(context.Background(), &stafiHubXRValidatorTypes.QueryCycleSecondsRequest{
			Denom: denom,
		})
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXRValidatorTypes.QueryCycleSecondsResponse), nil
}

func (c *Client) QueryShuffleSeconds(denom string) (*stafiHubXRValidatorTypes.QueryShuffleSecondsResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.Retry(func() (interface{}, error) {
		queryClient := stafiHubXRValidatorTypes.NewQueryClient(c.Ctx())
		return queryClient.ShuffleSeconds(context.Background(), &stafiHubXRValidatorTypes.QueryShuffleSecondsRequest{
			Denom: denom,
		})
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXRValidatorTypes.QueryShuffleSecondsResponse), nil
}

func (c *Client) QueryRValidatorList(denom, poolAddress string) (*stafiHubXRValidatorTypes.QueryRValidatorListResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.Retry(func() (interface{}, error) {
		queryClient := stafiHubXRValidatorTypes.NewQueryClient(c.Ctx())
		return queryClient.RValidatorList(context.Background(), &stafiHubXRValidatorTypes.QueryRValidatorListRequest{
			Denom:       denom,
			PoolAddress: poolAddress,
		})
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXRValidatorTypes.QueryRValidatorListResponse), nil
}

func (c *Client) QueryLatestVotedCycle(denom, poolAddress string) (*stafiHubXRValidatorTypes.QueryLatestVotedCycleResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.Retry(func() (interface{}, error) {
		queryClient := stafiHubXRValidatorTypes.NewQueryClient(c.Ctx())
		return queryClient.LatestVotedCycle(context.Background(), &stafiHubXRValidatorTypes.QueryLatestVotedCycleRequest{
			Denom:       denom,
			PoolAddress: poolAddress,
		})
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXRValidatorTypes.QueryLatestVotedCycleResponse), nil
}

func (c *Client) QueryLatestDealedCycle(denom, poolAddress string) (*stafiHubXRValidatorTypes.QueryLatestDealedCycleResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.Retry(func() (interface{}, error) {
		queryClient := stafiHubXRValidatorTypes.NewQueryClient(c.Ctx())
		return queryClient.LatestDealedCycle(context.Background(), &stafiHubXRValidatorTypes.QueryLatestDealedCycleRequest{
			Denom:       denom,
			PoolAddress: poolAddress,
		})
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXRValidatorTypes.QueryLatestDealedCycleResponse), nil
}

func (c *Client) QueryIcaPoolList(denom string) (*stafiHubXLedgerTypes.QueryIcaPoolListResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.Retry(func() (interface{}, error) {
		queryClient := stafiHubXLedgerTypes.NewQueryClient(c.Ctx())
		return queryClient.IcaPoolList(context.Background(), &stafiHubXLedgerTypes.QueryIcaPoolListRequest{
			Denom: denom,
		})
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXLedgerTypes.QueryIcaPoolListResponse), nil
}

func (c *Client) QueryInterchainTxStatus(propId string) (*stafiHubXLedgerTypes.QueryInterchainTxStatusResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.Retry(func() (interface{}, error) {
		queryClient := stafiHubXLedgerTypes.NewQueryClient(c.Ctx())
		return queryClient.InterchainTxStatus(context.Background(), &stafiHubXLedgerTypes.QueryInterchainTxStatusRequest{
			PropId: propId,
		})
	})
	if err != nil {
		return nil, err
	}
	return cc.(*stafiHubXLedgerTypes.QueryInterchainTxStatusResponse), nil
}
