package rpc_test

import (
	"context"
	"os"
	"testing"

	"github.com/b-harvest/gravity-dex-firestation/client/rpc"
	"github.com/b-harvest/gravity-dex-firestation/codec"

	"github.com/test-go/testify/require"
)

var (
	c *rpc.Client

	rpcAddress = "http://localhost:26657"
)

func TestMain(m *testing.M) {
	codec.SetCodec()

	c, _ = rpc.NewClient(rpcAddress, 5)

	os.Exit(m.Run())
}

func TestGetNetworkChainID(t *testing.T) {
	chainID, err := c.GetNetworkChainID(context.Background())
	require.NoError(t, err)

	t.Log(chainID)
}
