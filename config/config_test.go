package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/b-harvest/gravity-dex-firestation/config"
)

func TestReadConfigFile(t *testing.T) {
	configFilePath := "../example.toml"

	cfg, err := config.Read(configFilePath)
	require.NoError(t, err)

	require.Equal(t, "http://localhost:26657", cfg.RPC.Address)
	require.Equal(t, "localhost:9090", cfg.GRPC.Address)
	require.Equal(t, "YOUR_API_KEY", cfg.CoinMarketCap.APIKey)
}

func TestParseConfigString(t *testing.T) {
	var sampleConfig = `
[rpc]
address = "http://localhost:26657"

[grpc]
address = "localhost:9090"

[coinmarketcap]
api_key = "YOUR_API_KEY"

[wallet]
mnemonic = "<YOUR_MNEMONIC>"

[firestation]
cmc_symbol_a = "ATOM"
cmc_symbol_b = "LUNA"
denom_a = "uatom"
denom_b = "uluna"
`
	cfg, err := config.ParseString([]byte(sampleConfig))
	require.NoError(t, err)

	require.Equal(t, "http://localhost:26657", cfg.RPC.Address)
	require.Equal(t, "localhost:9090", cfg.GRPC.Address)
	require.Equal(t, "YOUR_API_KEY", cfg.CoinMarketCap.APIKey)
}
