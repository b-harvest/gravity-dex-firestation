<p align="center">
  <a href="https://github.com/b-harvest/gravity-dex-firestation" target="_blank"><img width="140" src="https://user-images.githubusercontent.com/20435620/117280451-92261580-ae9d-11eb-8907-f72a00320b22.jpeg" alt="B-Harvest"></a>
</p>

<h1 align="center">
    Gravity Dex Firestation 🚒
</h1>

## Overview

The pool price management bot to stabilize overpriced pools during the Gravity DEX incentivized testnet.
This project is developed solely for the purpose of pool price management during the incentivized testnet. 
It is developed in fast-pace; therfore it surely requires some work to generalize the codebase for better adoption and efficiency.
If you would like to use this for your own usage, make sure to go through the codebase thoroughly and change them as you need.

**Note**: Requires [Go 1.15+](https://golang.org/dl/)

## Version

- [Liquidity Module v1.2.5](https://github.com/tendermint/liquidity/tree/v1.2.5) 

## Configuration

This firestation repo requires a configuration file, `config.toml` in current working directory. An example of configuration file is available in `example.toml` and the config source code can be found in [here](./config.config.go).

## Build

```bash
# Clone the project 
git clone https://github.com/b-harvest/gravity-dex-firestation.git
cd gravity-dex-firestation

# At this point, CLI commands are not developed and all the logic is inside main function.
go run main.go
```

... 

## CoinMarketCap Symbols & Denoms

- [CoinMarketCap API Documentation](https://coinmarketcap.com/api/documentation/v1/)
- [Listings Latest](https://pro-api.coinmarketcap.com/v1/cryptocurrency/listings/latest?limit=5000): Your own API key is required to query this API.

| name | id | symbol | denom |
|---|---|---|---|
| Cosmos Hub    | 3794 | ATOM | uatom |
| Terra         | 4172 | LUNA | uluna |
| Akash Network | 7431 | AKT  | uakt  |
| IRISnet       | 3874 | IRIS | uiris |
| ... | ... |
