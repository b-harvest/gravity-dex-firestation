module github.com/b-harvest/gravity-dex-firestation

go 1.16

require (
	github.com/cosmos/cosmos-sdk v0.42.4
	github.com/cosmos/go-bip39 v1.0.0
	github.com/go-resty/resty/v2 v2.6.0
	github.com/pelletier/go-toml v1.9.0
	github.com/rs/zerolog v1.21.0
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/liquidity v1.2.5
	github.com/tendermint/tendermint v0.34.10
	github.com/test-go/testify v1.1.4
	google.golang.org/grpc v1.37.0
)

replace (
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
