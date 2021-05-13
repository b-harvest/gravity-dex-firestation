package util

import (
	"math/rand"
	"time"

	liqtypes "github.com/tendermint/liquidity/x/liquidity/types"
)

var CoinMarketCapMetadata = map[string]string{
	"uatom": "3794",
	"ubtsg": "8905",
	"udvpn": "2643",
	"uxprt": "7281",
	"uakt":  "7431",
	"uluna": "4172",
	"ungm":  "8279",
	"uiris": "3874",
}

// Shuffle randomizes liquidity pools.
func Shuffle(pools liqtypes.Pools) liqtypes.Pools {
	var cmcListedPools liqtypes.Pools

	for _, p := range pools {
		denomX := p.ReserveCoinDenoms[0]
		denomY := p.ReserveCoinDenoms[1]

		// remove the coins that are not listed in CoinMarketCap due to time limitation
		if CoinMarketCapMetadata[denomX] == "" || CoinMarketCapMetadata[denomY] == "" {
			continue
		}
		cmcListedPools = append(cmcListedPools, p)
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(cmcListedPools), func(i, j int) { cmcListedPools[i], cmcListedPools[j] = cmcListedPools[j], cmcListedPools[i] })

	return cmcListedPools
}

// Select picks n number of pools
func Select(pools liqtypes.Pools, n int) liqtypes.Pools {
	var r liqtypes.Pools
	for i := 0; i < n; i++ {
		r = append(r, pools[i])
	}
	return r
}
