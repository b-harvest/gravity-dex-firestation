package util

import (
	"math/rand"
	"time"

	liqtypes "github.com/tendermint/liquidity/x/liquidity/types"
)

// Shuffle randomizes liquidity pools.
func Shuffle(pools liqtypes.Pools) liqtypes.Pools {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(pools), func(i, j int) { pools[i], pools[j] = pools[j], pools[i] })
	return pools
}
