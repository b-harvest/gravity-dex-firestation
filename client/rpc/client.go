package rpc

import (
	"context"
	"fmt"

	rpcclient "github.com/tendermint/tendermint/rpc/client"
	rpc "github.com/tendermint/tendermint/rpc/client/http"
)

// Client wraps RPC client connection.
type Client struct {
	rpcclient.Client
}

// NewClient creates RPC client.
func NewClient(rpcURL string, timeout int64) (*Client, error) {
	rpcClient, err := rpc.NewWithTimeout(rpcURL, "/websocket", uint(5))
	if err != nil {
		return &Client{}, fmt.Errorf("failed to connect RPC client: %s", err)
	}

	return &Client{rpcClient}, nil
}

// GetNetworkChainID returns network chain id.
func (c *Client) GetNetworkChainID(ctx context.Context) (string, error) {
	status, err := c.Status(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get status: %v", err)
	}

	return status.NodeInfo.Network, nil
}
