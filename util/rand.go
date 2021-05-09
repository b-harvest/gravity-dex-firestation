package util

import (
	"math/rand"
	"time"

	"github.com/b-harvest/gravity-dex-firestation/types"
	liqtypes "github.com/tendermint/liquidity/x/liquidity/types"
)

// Shuffle randomizes liquidity pools.
func Shuffle(pools liqtypes.Pools) liqtypes.Pools {
	var cmcListedPools liqtypes.Pools

	// remove the coins that are not listed in CoinMarketCap due to time limitation
	for _, p := range pools {
		denomX := p.ReserveCoinDenoms[0]
		denomY := p.ReserveCoinDenoms[1]

		if types.CoinMarketCapMetadata[denomX] == "" || types.CoinMarketCapMetadata[denomY] == "" {
			continue
		}
		cmcListedPools = append(cmcListedPools, p)
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(cmcListedPools), func(i, j int) { cmcListedPools[i], cmcListedPools[j] = cmcListedPools[j], cmcListedPools[i] })

	return cmcListedPools
}

func Random(pools liqtypes.Pools, n int) liqtypes.Pools {
	var r liqtypes.Pools
	for i := 0; i < n; i++ {
		r = append(r, pools[i])
	}
	return r
}
