package client

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"syscall"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	xAuthTx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	xBankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	xDistriTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	xSlashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	xStakeTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stafihub/rtoken-relay-core/common/core"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

const retryLimit = 600
const waitTime = 2 * time.Second

// no 0x prefix
func (c *Client) QueryTxByHash(hashHexStr string) (*types.TxResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.retry(func() (interface{}, error) {
		return xAuthTx.QueryTx(c.Ctx(), hashHexStr)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*types.TxResponse), nil
}

func (c *Client) QueryDelegation(delegatorAddr types.AccAddress, validatorAddr types.ValAddress, height int64) (*xStakeTypes.QueryDelegationResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.retry(func() (interface{}, error) {
		client := c.Ctx().WithHeight(height)
		queryClient := xStakeTypes.NewQueryClient(client)
		params := &xStakeTypes.QueryDelegationRequest{
			DelegatorAddr: delegatorAddr.String(),
			ValidatorAddr: validatorAddr.String(),
		}
		return queryClient.Delegation(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*xStakeTypes.QueryDelegationResponse), nil
}

func (c *Client) QueryUnbondingDelegation(delegatorAddr types.AccAddress, validatorAddr types.ValAddress, height int64) (*xStakeTypes.QueryUnbondingDelegationResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.retry(func() (interface{}, error) {
		client := c.Ctx().WithHeight(height)
		queryClient := xStakeTypes.NewQueryClient(client)
		params := &xStakeTypes.QueryUnbondingDelegationRequest{
			DelegatorAddr: delegatorAddr.String(),
			ValidatorAddr: validatorAddr.String(),
		}
		return queryClient.UnbondingDelegation(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*xStakeTypes.QueryUnbondingDelegationResponse), nil
}

func (c *Client) QueryDelegations(delegatorAddr types.AccAddress, height int64) (*xStakeTypes.QueryDelegatorDelegationsResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.retry(func() (interface{}, error) {
		client := c.Ctx().WithHeight(height)
		queryClient := xStakeTypes.NewQueryClient(client)
		params := &xStakeTypes.QueryDelegatorDelegationsRequest{
			DelegatorAddr: delegatorAddr.String(),
			Pagination:    &query.PageRequest{},
		}
		return queryClient.DelegatorDelegations(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*xStakeTypes.QueryDelegatorDelegationsResponse), nil
}

func (c *Client) QueryDelegationRewards(delegatorAddr types.AccAddress, validatorAddr types.ValAddress, height int64) (*xDistriTypes.QueryDelegationRewardsResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.retry(func() (interface{}, error) {
		client := c.Ctx().WithHeight(height)
		queryClient := xDistriTypes.NewQueryClient(client)
		return queryClient.DelegationRewards(
			context.Background(),
			&xDistriTypes.QueryDelegationRewardsRequest{DelegatorAddress: delegatorAddr.String(), ValidatorAddress: validatorAddr.String()},
		)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*xDistriTypes.QueryDelegationRewardsResponse), nil
}

func (c *Client) QueryDelegationTotalRewards(delegatorAddr types.AccAddress, height int64) (*xDistriTypes.QueryDelegationTotalRewardsResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.retry(func() (interface{}, error) {
		client := c.Ctx().WithHeight(height)
		queryClient := xDistriTypes.NewQueryClient(client)
		return queryClient.DelegationTotalRewards(
			context.Background(),
			&xDistriTypes.QueryDelegationTotalRewardsRequest{DelegatorAddress: delegatorAddr.String()},
		)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*xDistriTypes.QueryDelegationTotalRewardsResponse), nil
}

func (c *Client) QueryBlock(height int64) (*ctypes.ResultBlock, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.retry(func() (interface{}, error) {
		node, err := c.Ctx().GetNode()
		if err != nil {
			return nil, err
		}
		return node.Block(context.Background(), &height)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*ctypes.ResultBlock), nil
}

func (c *Client) QueryAccount(addr types.AccAddress) (client.Account, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()
	return c.getAccount(0, addr)
}

func (c *Client) GetSequence(height int64, addr types.AccAddress) (uint64, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()
	account, err := c.getAccount(height, addr)
	if err != nil {
		return 0, err
	}
	return account.GetSequence(), nil
}

func (c *Client) QueryBalance(addr types.AccAddress, denom string, height int64) (*xBankTypes.QueryBalanceResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.retry(func() (interface{}, error) {
		client := c.Ctx().WithHeight(height)
		queryClient := xBankTypes.NewQueryClient(client)
		params := xBankTypes.NewQueryBalanceRequest(addr, denom)
		return queryClient.Balance(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*xBankTypes.QueryBalanceResponse), nil
}

func (c *Client) QuerySupplyOf(denom string, height int64) (*xBankTypes.QuerySupplyOfResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.retry(func() (interface{}, error) {
		client := c.Ctx().WithHeight(height)
		queryClient := xBankTypes.NewQueryClient(client)
		params := xBankTypes.QuerySupplyOfRequest{Denom: denom}
		return queryClient.SupplyOf(context.Background(), &params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*xBankTypes.QuerySupplyOfResponse), nil
}

func (c *Client) GetCurrentBlockHeight() (int64, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	status, err := c.getStatus()
	if err != nil {
		return 0, err
	}
	return status.SyncInfo.LatestBlockHeight, nil
}

func (c *Client) getStatus() (*ctypes.ResultStatus, error) {
	cc, err := c.retry(func() (interface{}, error) {
		return c.Ctx().Client.Status(context.Background())
	})
	if err != nil {
		return nil, err
	}
	return cc.(*ctypes.ResultStatus), nil
}

func (c *Client) GetAccount() (client.Account, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	return c.getAccount(0, c.Ctx().FromAddress)
}

func (c *Client) getAccount(height int64, addr types.AccAddress) (client.Account, error) {
	cc, err := c.retry(func() (interface{}, error) {
		client := c.Ctx().WithHeight(height)
		return client.AccountRetriever.GetAccount(c.Ctx(), addr)
	})
	if err != nil {
		return nil, err
	}
	return cc.(client.Account), nil
}

func (c *Client) GetTxs(events []string, page, limit int, orderBy string) (*types.SearchTxsResult, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.retry(func() (interface{}, error) {
		return xAuthTx.QueryTxsByEvents(c.Ctx(), events, page, limit, orderBy)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*types.SearchTxsResult), nil
}

func (c *Client) GetTxsWithParseErrSkip(events []string, page, limit int, orderBy string) (*types.SearchTxsResult, int, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	externalSkipCount := 0
	cc, err := c.retry(func() (interface{}, error) {
		result, skip, err := xAuthTx.QueryTxsByEventsWithParseErrSkip(c.Ctx(), events, page, limit, orderBy)
		externalSkipCount = skip
		return result, err
	})
	if err != nil {
		return nil, 0, err
	}
	return cc.(*types.SearchTxsResult), externalSkipCount, nil
}

// will skip txs that parse failed
func (c *Client) GetBlockTxsWithParseErrSkip(height int64) ([]*types.TxResponse, error) {
	// tendermint max limit 100
	txs := make([]*types.TxResponse, 0)
	limit := 50
	initPage := 1
	totalSkipCount := 0
	searchTxs, skipCount, err := c.GetTxsWithParseErrSkip([]string{fmt.Sprintf("tx.height=%d", height)}, initPage, limit, "asc")
	if err != nil {
		return nil, err
	}
	totalSkipCount += skipCount
	txs = append(txs, searchTxs.Txs...)
	for page := initPage + 1; page <= int(searchTxs.PageTotal); page++ {
		subSearchTxs, skipCount, err := c.GetTxsWithParseErrSkip([]string{fmt.Sprintf("tx.height=%d", height)}, page, limit, "asc")
		if err != nil {
			return nil, err
		}
		totalSkipCount += skipCount
		txs = append(txs, subSearchTxs.Txs...)
	}

	if int(searchTxs.TotalCount) != len(txs)+totalSkipCount {
		return nil, fmt.Errorf("tx total count overflow, searchTxs.TotalCount: %d txs len: %d", searchTxs.TotalCount, len(txs)+totalSkipCount)
	}
	return txs, nil
}

func (c *Client) GetBlockTxs(height int64) ([]*types.TxResponse, error) {
	// tendermint max limit 100
	txs := make([]*types.TxResponse, 0)
	limit := 50
	initPage := 1
	searchTxs, err := c.GetTxs([]string{fmt.Sprintf("tx.height=%d", height)}, initPage, limit, "asc")
	if err != nil {
		return nil, err
	}
	txs = append(txs, searchTxs.Txs...)
	for page := initPage + 1; page <= int(searchTxs.PageTotal); page++ {
		subSearchTxs, err := c.GetTxs([]string{fmt.Sprintf("tx.height=%d", height)}, page, limit, "asc")
		if err != nil {
			return nil, err
		}
		txs = append(txs, subSearchTxs.Txs...)
	}

	if int(searchTxs.TotalCount) != len(txs) {
		return nil, fmt.Errorf("tx total count overflow, searchTxs.TotalCount: %d txs len: %d", searchTxs.TotalCount, len(txs))
	}
	return txs, nil
}

func (c *Client) GetBlockResults(height int64) (*ctypes.ResultBlockResults, error) {
	cc, err := c.retry(func() (interface{}, error) {
		return c.clientCtx.Client.BlockResults(context.Background(), &height)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*ctypes.ResultBlockResults), nil
}

func (c *Client) GetChainId() (string, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	status, err := c.getStatus()
	if err != nil {
		return "", err
	}
	return status.NodeInfo.Network, nil
}

func (c *Client) QueryBondedDenom() (*xStakeTypes.QueryParamsResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.retry(func() (interface{}, error) {
		queryClient := xStakeTypes.NewQueryClient(c.Ctx())
		params := xStakeTypes.QueryParamsRequest{}
		return queryClient.Params(context.Background(), &params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*xStakeTypes.QueryParamsResponse), nil
}

func (c *Client) QueryValidators(height int64) (*xStakeTypes.QueryValidatorsResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.retry(func() (interface{}, error) {
		client := c.Ctx().WithHeight(height)
		queryClient := xStakeTypes.NewQueryClient(client)
		params := &xStakeTypes.QueryValidatorsRequest{
			Pagination: &query.PageRequest{
				Offset:     0,
				Limit:      1000,
				CountTotal: false,
				Reverse:    false,
			},
		}
		return queryClient.Validators(context.Background(), params)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*xStakeTypes.QueryValidatorsResponse), nil
}

func (c *Client) QuerySigningInfo(consAddr string, height int64) (*xSlashingTypes.QuerySigningInfoResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.retry(func() (interface{}, error) {
		queryClient := xSlashingTypes.NewQueryClient(c.Ctx().WithHeight(height))
		return queryClient.SigningInfo(
			context.Background(),
			&xSlashingTypes.QuerySigningInfoRequest{
				ConsAddress: consAddr,
			},
		)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*xSlashingTypes.QuerySigningInfoResponse), nil
}

func (c *Client) QueryAllSigningInfos(height int64) (*xSlashingTypes.QuerySigningInfosResponse, error) {
	done := core.UseSdkConfigContext(GetAccountPrefix())
	defer done()

	cc, err := c.retry(func() (interface{}, error) {
		queryClient := xSlashingTypes.NewQueryClient(c.Ctx().WithHeight(height))
		return queryClient.SigningInfos(
			context.Background(),
			&xSlashingTypes.QuerySigningInfosRequest{},
		)
	})
	if err != nil {
		return nil, err
	}
	return cc.(*xSlashingTypes.QuerySigningInfosResponse), nil
}

func (c *Client) Retry(f func() (interface{}, error)) (interface{}, error) {
	return c.retry(f)
}

// only retry func when return connection err here
func (c *Client) retry(f func() (interface{}, error)) (interface{}, error) {
	var err error
	var result interface{}
	for i := 0; i < retryLimit; i++ {
		result, err = f()
		if err != nil {
			c.logger.Debug("retry",
				"endpoint index", c.CurrentEndpointIndex(),
				"err", err)
			// connection err case
			if isConnectionError(err) {
				c.ChangeEndpoint()
				time.Sleep(waitTime)
				continue
			}
			// business err case or other err case not captured
			for j := 0; j < len(c.rpcClientList)*2; j++ {
				c.ChangeEndpoint()
				subResult, subErr := f()

				if subErr != nil {
					c.logger.Debug("retry",
						"endpoint index", c.CurrentEndpointIndex(),
						"subErr", err)
					// filter connection err
					if isConnectionError(subErr) {
						continue
					}

					result = subResult
					err = subErr
					continue
				}

				result = subResult
				err = subErr
				// if ok when using this rpc, just return
				return result, err
			}

			// still failed after try all rpc, just return err
			return result, err

		}
		// no err, just return
		return result, err
	}
	return nil, fmt.Errorf("reach retry limit. err: %s", err)
}

func isConnectionError(err error) bool {
	switch t := err.(type) {
	case *url.Error:
		if t.Timeout() || t.Temporary() {
			return true
		}
		return isConnectionError(t.Err)
	}

	switch t := err.(type) {
	case *net.OpError:
		if t.Op == "dial" || t.Op == "read" {
			return true
		}
		return isConnectionError(t.Err)

	case syscall.Errno:
		if t == syscall.ECONNREFUSED {
			return true
		}
	}

	switch t := err.(type) {
	case wrapError:
		newErr := t.Unwrap()
		return isConnectionError(newErr)
	}

	if err != nil {
		// json unmarshal err when rpc server shutting down
		if strings.Contains(err.Error(), "looking for beginning of value") {
			return true
		}
		// server goroutine panic
		if strings.Contains(err.Error(), "recovered") {
			return true
		}
		if strings.Contains(err.Error(), "panic") {
			return true
		}
		if strings.Contains(err.Error(), "Internal server error") {
			return true
		}
	}

	return false
}

type wrapError interface {
	Unwrap() error
}
