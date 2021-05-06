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
	LCD           LCDConfig           `toml:"lcd"`
	CoinMarketCap CoinMarketCapConfig `toml:"coinmarketcap"`
}

// DefaultRPCConfig is the default RPC config.
var DefaultRPCConfig = RPCConfig{
	Address: "http://localhost:26657",
}

// RPCConfig contains the configuration of the RPC endpoint.
type RPCConfig struct {
	Address string `toml:"address"`
}

// DefaultGRPCConfig is the default gRPC config.
var DefaultGRPCConfig = GRPCConfig{
	Address: "localhost:9090",
}

// GRPCConfig contains the configuration of the gRPC endpoint.
type GRPCConfig struct {
	Address string `toml:"address"`
}

// DefaultLCDConfig is the default REST API endpoint config.
var DefaultLCDConfig = LCDConfig{
	Address: "http://localhost:1317",
}

// LCDConfig contains the configuration of the REST server endpoint.
type LCDConfig struct {
	Address string `toml:"address"`
}

// DefaultCoinMarketCapConfig is the default CoinMarketCap config.
var DefaultCoinMarketCapConfig = CoinMarketCapConfig{
	APIKey: "",
}

// CoinMarketCapConfig contains the API key to request CoinMarketCap's APIs.
type CoinMarketCapConfig struct {
	APIKey string `toml:"api_key"`
}

// DefaultConfig returns default Config object.
func DefaultConfig() Config {
	return Config{
		RPC:           DefaultRPCConfig,
		GRPC:          DefaultGRPCConfig,
		LCD:           DefaultLCDConfig,
		CoinMarketCap: DefaultCoinMarketCapConfig,
	}
}

// NewConfig builds a new Config object.
func NewConfig(rpc RPCConfig, gRPC GRPCConfig, lcd LCDConfig, coinmarketcap CoinMarketCapConfig) Config {
	return Config{
		RPC:           rpc,
		GRPC:          gRPC,
		LCD:           lcd,
		CoinMarketCap: coinmarketcap,
	}
}

// SetupConfig takes the path to a configuration file and returns the properly parsed configuration.
func Read(configPath string) (Config, error) {
	if configPath == "" {
		return Config{}, fmt.Errorf("empty configuration path")
	}

	log.Debug().Msg("reading config file")

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

	log.Debug().Msg("parsing config data")

	err := toml.Unmarshal(configData, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("failed to decode config: %s", err)
	}

	return cfg, nil
}
