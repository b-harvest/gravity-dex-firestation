package tx

import (
	"context"
	"fmt"

	"github.com/b-harvest/gravity-dex-firestation/client"

	liqtypes "github.com/tendermint/liquidity/x/liquidity/types"

	sdkclientx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

var (
	gasLimit = uint64(100000000)
	memo     = ""
)

// Transaction is an object that has common fields when signing transaction.
type Transaction struct {
	Client  *client.Client `json:"client"`
	ChainID string         `json:"chain_id"`
	Fees    sdk.Coins      `json:"fees"`
}

// NewTransaction returns new Transaction object.
func NewTransaction(client *client.Client, chainID string, fees sdk.Coins) *Transaction {
	return &Transaction{
		Client:  client,
		ChainID: chainID,
		Fees:    fees,
	}
}

// MsgSwap creates swap message and returns MsgWithdraw MsgSwap message.
func MsgSwap(poolCreator string, poolId uint64, swapTypeId uint32, offerCoin sdk.Coin,
	demandCoinDenom string, orderPrice sdk.Dec, swapFeeRate sdk.Dec) (sdk.Msg, error) {
	accAddr, err := sdk.AccAddressFromBech32(poolCreator)
	if err != nil {
		return &liqtypes.MsgSwapWithinBatch{}, err
	}

	msg := liqtypes.NewMsgSwapWithinBatch(accAddr, poolId, swapTypeId, offerCoin, demandCoinDenom, orderPrice, swapFeeRate)

	if err := msg.ValidateBasic(); err != nil {
		return &liqtypes.MsgSwapWithinBatch{}, err
	}

	return msg, nil
}

// Sign signs message(s) with the account's private key and braodacasts the message(s).
func (t *Transaction) Sign(ctx context.Context, accSeq uint64, accNum uint64, privKey *secp256k1.PrivKey, msgs ...sdk.Msg) ([]byte, error) {
	txBuilder := t.Client.CliCtx.TxConfig.NewTxBuilder()
	txBuilder.SetMsgs(msgs...)
	txBuilder.SetGasLimit(gasLimit)
	txBuilder.SetFeeAmount(t.Fees)
	txBuilder.SetMemo(memo)

	signMode := t.Client.CliCtx.TxConfig.SignModeHandler().DefaultMode()

	sigV2 := signing.SignatureV2{
		PubKey: privKey.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  signMode,
			Signature: nil,
		},
		Sequence: accSeq,
	}

	err := txBuilder.SetSignatures(sigV2)
	if err != nil {
		return nil, fmt.Errorf("failed to set signatures: %s", err)
	}

	signerData := authsigning.SignerData{
		ChainID:       t.ChainID,
		AccountNumber: accNum,
		Sequence:      accSeq,
	}

	sigV2, err = sdkclientx.SignWithPrivKey(signMode, signerData, txBuilder, privKey, t.Client.CliCtx.TxConfig, accSeq)
	if err != nil {
		return nil, fmt.Errorf("failed to sign with private key: %s", err)
	}

	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		return nil, fmt.Errorf("failed to set signatures: %s", err)
	}

	txByte, err := t.Client.CliCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, fmt.Errorf("failed to encode tx and get raw tx data: %s", err)
	}

	return txByte, nil
}

// BroadcastTx broadcasts transaction.
func (t *Transaction) BroadcastTx(ctx context.Context, txBytes []byte) (*sdktx.BroadcastTxResponse, error) {
	client := t.Client.GRPC.GetTxClient()

	req := &sdktx.BroadcastTxRequest{
		TxBytes: txBytes,
		Mode:    sdktx.BroadcastMode_BROADCAST_MODE_BLOCK,
	}
	return client.BroadcastTx(ctx, req)
}
