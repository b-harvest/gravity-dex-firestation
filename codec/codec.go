package codec

import (
	"github.com/cosmos/cosmos-sdk/codec"

	liquidityapp "github.com/tendermint/liquidity/app"
	"github.com/tendermint/liquidity/app/params"
)

// Codec is the application-wide Amino codec and is initialized upon package loading.
var (
	AppCodec       codec.Marshaler
	AminoCodec     *codec.LegacyAmino
	EncodingConfig params.EncodingConfig
)

// SetCodec sets encoding config.
func SetCodec() {
	EncodingConfig = liquidityapp.MakeEncodingConfig()
	AppCodec = EncodingConfig.Marshaler
	AminoCodec = EncodingConfig.Amino
}
