package wallet_test

import (
	"testing"

	bip39 "github.com/cosmos/go-bip39"

	"github.com/test-go/testify/require"

	"github.com/b-harvest/gravity-dex-firestation/wallet"
)

func TestRecoverAccAddrFromMnemonic(t *testing.T) {
	testCases := []struct {
		mnemonic   string
		password   string
		expAccAddr string
	}{
		{
			mnemonic:   "guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host",
			password:   "",
			expAccAddr: "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
		},
		{
			mnemonic:   "friend excite rough reopen cover wheel spoon convince island path clean monkey play snow number walnut pull lock shoot hurry dream divide concert discover",
			password:   "",
			expAccAddr: "cosmos1mzgucqnfr2l8cj5apvdpllhzt4zeuh2cshz5xu",
		},
		{
			mnemonic:   "fuel obscure melt april direct second usual hair leave hobby beef bacon solid drum used law mercy worry fat super must ritual bring faculty",
			password:   "",
			expAccAddr: "cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny",
		},
	}

	for _, tc := range testCases {
		accAddr, _, err := wallet.RecoverAccountFromMnemonic(tc.mnemonic, tc.password)
		require.NoError(t, err)

		require.Equal(t, tc.expAccAddr, accAddr)
	}
}

func TestNewMnemonic(t *testing.T) {
	for i := 0; i < 5; i++ {
		entropy, err := bip39.NewEntropy(256)
		require.NoError(t, err)

		mnemonic, err := bip39.NewMnemonic(entropy)
		require.NoError(t, err)

		accAddr, _, err := wallet.RecoverAccountFromMnemonic(mnemonic, "")
		require.NoError(t, err)

		t.Log(mnemonic, accAddr)
	}
}
