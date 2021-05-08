package grpc_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/b-harvest/gravity-dex-firestation/client/grpc"
	"github.com/b-harvest/gravity-dex-firestation/codec"
	"github.com/b-harvest/gravity-dex-firestation/config"

	"github.com/test-go/testify/require"
)

var (
	c *grpc.Client

	grpcAddress = "localhost:9090"
)

func TestMain(m *testing.M) {
	codec.SetCodec()

	c, _ = grpc.NewClient(grpcAddress, config.DefaultCoinMarketCapConfig)

	os.Exit(m.Run())
}

func TestPoolReserves(t *testing.T) {
	ctx := context.Background()

	reservePoolDenoms := []string{"uatom", "uluna"}

	reserveA, reserveB, err := c.GetPoolReserves(ctx, reservePoolDenoms)
	require.NoError(t, err)

	fmt.Println("reserveA: ", reserveA)
	fmt.Println("reserveB: ", reserveB)
}
