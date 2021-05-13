package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/b-harvest/gravity-dex-firestation/config"

	liqtypes "github.com/tendermint/liquidity/x/liquidity/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// Client wraps GRPC client connection.
type Client struct {
	client *grpc.ClientConn
	cfg    config.CoinMarketCapConfig
}

// NewClient creates GRPC client.
func NewClient(grpcURL string, cfg config.CoinMarketCapConfig) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := grpc.DialContext(ctx, grpcURL, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return &Client{}, fmt.Errorf("failed to connect GRPC client: %s", err)
	}

	return &Client{
		client: client,
		cfg:    cfg,
	}, nil
}

// IsNotFound returns not found status.
func IsNotFound(err error) bool {
	return status.Convert(err).Code() == codes.NotFound
}

// GetAllBalances returns all account balances.
func (c *Client) GetAllBalances(ctx context.Context, address string) (sdk.Coins, error) {
	bankClient := banktypes.NewQueryClient(c.client)

	req := banktypes.QueryAllBalancesRequest{
		Address: address,
	}

	resp, err := bankClient.AllBalances(ctx, &req)
	if err != nil {
		return sdk.Coins{}, err
	}

	return resp.GetBalances(), nil
}

// GetBaseAccountInfo returns base account information.
func (c *Client) GetBaseAccountInfo(ctx context.Context, address string) (authtypes.BaseAccount, error) {
	client := authtypes.NewQueryClient(c.client)

	req := authtypes.QueryAccountRequest{
		Address: address,
	}

	resp, err := client.Account(ctx, &req)
	if err != nil {
		return authtypes.BaseAccount{}, err
	}

	var acc authtypes.BaseAccount
	err = acc.Unmarshal(resp.GetAccount().Value)
	if err != nil {
		return authtypes.BaseAccount{}, err
	}

	return acc, nil
}

// GetPoolReserves returns pool reserves of the pool.
func (c *Client) GetPoolReserves(ctx context.Context, reservePoolDenoms []string) (sdk.Dec, sdk.Dec, error) {
	poolName := liqtypes.PoolName(reservePoolDenoms, 1)
	reserveAcc := liqtypes.GetPoolReserveAcc(poolName)

	balances, err := c.GetAllBalances(ctx, reserveAcc.String())
	if err != nil {
		return sdk.ZeroDec(), sdk.ZeroDec(), fmt.Errorf("failed to get reserve account balances: %s", err)
	}

	amountX := sdk.ZeroInt()
	amountY := sdk.ZeroInt()

	if balances.IsValid() {
		for _, b := range balances {
			if b.GetDenom() == reservePoolDenoms[0] {
				amountX = b.Amount
			}
			if b.GetDenom() == reservePoolDenoms[1] {
				amountY = b.Amount
			}
		}
	}

	return amountX.ToDec(), amountY.ToDec(), nil
}

// GetPool returns pool information.
func (c *Client) GetPool(ctx context.Context, poolId uint64) (liqtypes.Pool, error) {
	client := c.GetLiquidityQueryClient()

	req := liqtypes.QueryLiquidityPoolRequest{
		PoolId: poolId,
	}

	resp, err := client.LiquidityPool(ctx, &req)
	if err != nil {
		return liqtypes.Pool{}, err
	}

	return resp.GetPool(), nil
}

// GetAllPools returns all existing pools.
func (c *Client) GetAllPools(ctx context.Context) (liqtypes.Pools, error) {
	client := c.GetLiquidityQueryClient()

	req := liqtypes.QueryLiquidityPoolsRequest{}

	resp, err := client.LiquidityPools(ctx, &req)
	if err != nil {
		return liqtypes.Pools{}, err
	}

	return resp.GetPools(), nil
}

// GetLiquidityQueryClient returns a object of queryClient
func (c *Client) GetLiquidityQueryClient() liqtypes.QueryClient {
	return liqtypes.NewQueryClient(c.client)
}

// GetTxClient returns an object of service client.
func (c *Client) GetTxClient() sdktx.ServiceClient {
	return sdktx.NewServiceClient(c.client)
}
