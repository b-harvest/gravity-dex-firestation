package wallet

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	bip39 "github.com/cosmos/go-bip39"
)

// RecoverAccountFromMnemonic recovers private key from mnemonic and return account address after bech32 encoding.
func RecoverAccountFromMnemonic(mnemonic string, password string) (string, *secp256k1.PrivKey, error) {
	seed := bip39.NewSeed(mnemonic, password)
	masterKey, ch := hd.ComputeMastersFromSeed(seed)
	priv, err := hd.DerivePrivateKeyForPath(masterKey, ch, sdktypes.GetConfig().GetFullFundraiserPath()) // "44'/118'/0'/0/0"
	if err != nil {
		return "", &secp256k1.PrivKey{}, fmt.Errorf("failed to derive private key for path: %s", err)
	}

	privKey := &secp256k1.PrivKey{Key: priv}
	pubKey := privKey.PubKey()

	accAddr, err := bech32.ConvertAndEncode(sdktypes.GetConfig().GetBech32AccountAddrPrefix(), pubKey.Address())
	if err != nil {
		return "", &secp256k1.PrivKey{}, fmt.Errorf("failed to convert and encode address: %s", err)
	}

	return accAddr, privKey, nil
}
