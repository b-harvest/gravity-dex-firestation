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
	// go clean -testcache
	testCases := []struct {
		name            string
		reserCoinDenoms []string
	}{

		{
			"",
			[]string{"udvpn", "ungm"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reserveA, reserveB, err := c.GetPoolReserves(context.Background(), tc.reserCoinDenoms)
			require.NoError(t, err)

			fmt.Printf("denomA: %s reserveA: %s \n", tc.reserCoinDenoms[0], reserveA)
			fmt.Printf("denomB: %s reserveB: %s \n", tc.reserCoinDenoms[1], reserveB)
			fmt.Println("")
		})
	}
}

func TestAllPools(t *testing.T) {
	pools, err := c.GetAllPools(context.Background())
	require.NoError(t, err)

	for _, p := range pools {
		fmt.Println(p)
	}

	fmt.Println(len(pools))
}
