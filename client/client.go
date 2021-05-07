package client

import (
	"github.com/b-harvest/gravity-dex-firestation/client/clictx"
	"github.com/b-harvest/gravity-dex-firestation/client/grpc"
	"github.com/b-harvest/gravity-dex-firestation/client/market"
	"github.com/b-harvest/gravity-dex-firestation/client/rpc"
	"github.com/b-harvest/gravity-dex-firestation/codec"
	"github.com/b-harvest/gravity-dex-firestation/config"
)

// Client is a wrapper for various clients.
type Client struct {
	CliCtx *clictx.Client
	RPC    *rpc.Client
	GRPC   *grpc.Client
	Market *market.Client
}

// NewClient creates a new Client with the given configuration.
func NewClient(rpcURL string, grpcURL string, cmcConfig config.CoinMarketCapConfig) (*Client, error) {
	codec.SetCodec()

	rpcClient, err := rpc.NewClient(rpcURL, 5)
	if err != nil {
		return &Client{}, err
	}

	grpcClient, err := grpc.NewClient(grpcURL, cmcConfig)
	if err != nil {
		return &Client{}, err
	}

	cliCtx := clictx.NewClient(rpcURL, rpcClient.Client)

	marketClient := market.NewClient(cmcConfig)

	return &Client{
		CliCtx: cliCtx,
		RPC:    rpcClient,
		GRPC:   grpcClient,
		Market: marketClient,
	}, nil
}

// GetRPCClient returns RPC client.
func (c *Client) GetRPCClient() *rpc.Client {
	return c.RPC
}

// GetGRPCClient returns GRPC client.
func (c *Client) GetGRPCClient() *grpc.Client {
	return c.GRPC
}

// GetMarketClient returns Market client.
func (c *Client) GetMarketClient() *market.Client {
	return c.Market
}
