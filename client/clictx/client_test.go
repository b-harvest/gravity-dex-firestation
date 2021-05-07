package clictx_test

import (
	"os"
	"testing"

	"github.com/b-harvest/gravity-dex-firestation/client/clictx"
	"github.com/b-harvest/gravity-dex-firestation/client/rpc"
	"github.com/b-harvest/gravity-dex-firestation/codec"

	"github.com/test-go/testify/require"
)

var (
	c *clictx.Client

	rpcAddress = "http://localhost:26657"
)

func TestMain(m *testing.M) {
	codec.SetCodec()

	rpcClient, _ := rpc.NewClient(rpcAddress, 5)

	c = clictx.NewClient(rpcAddress, rpcClient)

	os.Exit(m.Run())
}

func TestGetAccount(t *testing.T) {
	address := ""
	acc, err := c.GetAccount(address)
	require.NoError(t, err)

	t.Log("pubkey :", acc.GetPubKey())
}
