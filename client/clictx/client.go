package clictx

import (
	"github.com/b-harvest/gravity-dex-firestation/codec"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

// Client wraps Cosmos SDK client context.
type Client struct {
	sdkclient.Context
}

// NewClient creates Cosmos SDK client.
func NewClient(rpcURL string, rpcClient rpcclient.Client) *Client {
	cliCtx := sdkclient.Context{}.
		WithNodeURI(rpcURL).
		WithClient(rpcClient).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithJSONMarshaler(codec.EncodingConfig.Marshaler).
		WithLegacyAmino(codec.EncodingConfig.Amino).
		WithTxConfig(codec.EncodingConfig.TxConfig).
		WithInterfaceRegistry(codec.EncodingConfig.InterfaceRegistry)

	return &Client{cliCtx}
}

// GetAccount checks account type and returns account interface.
func (c *Client) GetAccount(address string) (sdkclient.Account, error) {
	accAddr, err := sdktypes.AccAddressFromBech32(address)
	if err != nil {
		return nil, err
	}

	ar := authtypes.AccountRetriever{}

	acc, _, err := ar.GetAccountWithHeight(c.Context, accAddr)
	if err != nil {
		return nil, err
	}

	return acc, nil
}
