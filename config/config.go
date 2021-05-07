package config

import (
	"fmt"
	"io/ioutil"

	"github.com/pelletier/go-toml"

	"github.com/rs/zerolog/log"
)

var (
	DefaultConfigPath = "./config.toml"
)

// Config defines all necessary configuration parameters.
type Config struct {
	RPC           RPCConfig           `toml:"rpc"`
	GRPC          GRPCConfig          `toml:"grpc"`
	Wallet        WalletConfig        `toml:"wallet"`
	CoinMarketCap CoinMarketCapConfig `toml:"coinmarketcap"`
	FireStation   FireStationConfig   `toml:"firestation"`
}

// DefaultRPCConfig is the default RPCConfig.
var DefaultRPCConfig = RPCConfig{
	Address: "http://localhost:26657",
}

// RPCConfig contains the configuration of the RPC endpoint.
type RPCConfig struct {
	Address string `toml:"address"`
}

// DefaultGRPCConfig is the default GRPCConfig.
var DefaultGRPCConfig = GRPCConfig{
	Address: "localhost:9090",
}

// GRPCConfig contains the configuration of the gRPC endpoint.
type GRPCConfig struct {
	Address string `toml:"address"`
}

// DefaultCoinMarketCapConfig is the default CoinMarketCap.
var DefaultCoinMarketCapConfig = CoinMarketCapConfig{
	APIKey: "",
}

// CoinMarketCapConfig contains the API key to request CoinMarketCap's APIs.
type CoinMarketCapConfig struct {
	APIKey string `toml:"api_key"`
}

// DefaultFireStationConfig is the default FireStationConfig.
var DefaultFireStationConfig = FireStationConfig{
	CmcIdA: "3794", // ATOM
	CmcIdB: "4172", // LUNA
	DenomA: "uatom",
	DenomB: "uluna",
}

// FireStationConfig contains two different denoms and CoinMarketCap symbols.
type FireStationConfig struct {
	CmcIdA string `toml:"cmc_id_a"`
	CmcIdB string `toml:"cmc_id_b"`
	DenomA string `toml:"denom_a"`
	DenomB string `toml:"denom_b"`
}

// DefaultWalletConfig is the default WalletConfig.
var DefaultWalletConfig = WalletConfig{
	Mnemonic: "",
}

// WalletConfig contains mnemonic which should have sufficient balances to stabilize pool price.
type WalletConfig struct {
	Mnemonic string `toml:"mnemonic"`
}

// DefaultConfig returns default Config object.
func DefaultConfig() Config {
	return Config{
		RPC:           DefaultRPCConfig,
		GRPC:          DefaultGRPCConfig,
		CoinMarketCap: DefaultCoinMarketCapConfig,
		FireStation:   DefaultFireStationConfig,
	}
}

// NewConfig builds a new Config object.
func NewConfig(rpc RPCConfig, gRPC GRPCConfig, coinmarketcap CoinMarketCapConfig) Config {
	return Config{
		RPC:           rpc,
		GRPC:          gRPC,
		CoinMarketCap: coinmarketcap,
	}
}

// SetupConfig takes the path to a configuration file and returns the properly parsed configuration.
func Read(configPath string) (Config, error) {
	if configPath == "" {
		return Config{}, fmt.Errorf("empty configuration path")
	}

	log.Debug().Msg("reading config file...")

	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config: %s", err)
	}

	return ParseString(configData)
}

// ParseString attempts to read and parse  config from the given string bytes.
// An error reading or parsing the config results in a panic.
func ParseString(configData []byte) (Config, error) {
	var cfg Config

	log.Debug().Msg("parsing config data...")

	err := toml.Unmarshal(configData, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("failed to decode config: %s", err)
	}

	return cfg, nil
}
