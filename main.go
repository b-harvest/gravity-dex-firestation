package main

import (
	"context"
	"fmt"
	"os"

	"github.com/b-harvest/gravity-dex-firestation/client"
	"github.com/b-harvest/gravity-dex-firestation/config"
	"github.com/b-harvest/gravity-dex-firestation/tx"
	"github.com/b-harvest/gravity-dex-firestation/wallet"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// human-readable pretty logging is the default logging format
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := config.Read(config.DefaultConfigPath)
	if err != nil {
		fmt.Printf("failed to read config: %s", err)
		return
	}

	client, err := client.NewClient(cfg.RPC.Address, cfg.GRPC.Address, cfg.CoinMarketCap)
	if err != nil {
		fmt.Printf("failed to create new config: %s", err)
		return
	}

	stablizePoolPrice(cfg, client)
}

// stablizePoolPrice stablizes pool price of ATOM/LUNA pool.
// This provides users arbitrage opportunity for overpriced luna by managing the pool price.
func stablizePoolPrice(cfg config.Config, client *client.Client) error {
	log.Info().Msg("stablizing pool price...")

	ctx := context.Background()

	chainID, err := client.RPC.GetNetworkChainID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain id: %s", err)
	}

	accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(cfg.Wallet.Mnemonic, "")
	if err != nil {
		return fmt.Errorf("failed to retrieve account and private key from mnemonic: %s", err)
	}

	account, err := client.GRPC.GetBaseAccountInfo(ctx, accAddr)
	if err != nil {
		return fmt.Errorf("failed to get account information: %s", err)
	}

	accSeq := account.GetSequence()
	accNum := account.GetAccountNumber()

	reservePoolDenoms := []string{cfg.FireStation.DenomA, cfg.FireStation.DenomB}
	cmcIds := []string{cfg.FireStation.CmcIdA, cfg.FireStation.CmcIdB}

	reserveAmtX, reserveAmtY, err := client.GRPC.GetPoolReserves(ctx, reservePoolDenoms)
	if err != nil {
		return fmt.Errorf("failed to get pool price: %s", err)
	}

	globalPriceX, globalPriceY, err := client.Market.GetMarketPrices(ctx, cmcIds)
	if err != nil {
		return fmt.Errorf("failed to get pool prices: %s", err)
	}

	poolPrice := reserveAmtX.Quo(reserveAmtY)                  // POOLPRICE   = ATOMRESERVE/LUNARESERVE
	globalPrice := globalPriceX.Quo(globalPriceY)              // GLOBALPRICE = ATOMUSD/LUNAUSD
	priceDiff := globalPrice.Quo(poolPrice).Sub(sdk.NewDec(1)) // PRICEDIFF   = GLOBALPRICE/POOLPRICE - 1

	log.Debug().
		Str("reserveAmountX", reserveAmtX.String()).
		Str("reserveAmountY", reserveAmtY.String()).
		Str("reservePoolPrice", poolPrice.String()).
		Str("globalPriceX", globalPriceX.String()).
		Str("globalPriceY", globalPriceY.String()).
		Str("globalPrice", globalPrice.String()).
		Str("priceDiff", priceDiff.String()).
		Msg("")

	transaction := tx.NewTransaction(client, chainID)

	switch {
	// LUNA is overpriced / ATOM is underpriced / price diff is greater than 10%
	case priceDiff.IsPositive() && priceDiff.GT(sdk.NewDecWithPrec(1, 1)):
		log.Info().Msgf("priceDiff is positive; selling '%s' buying '%s'", reservePoolDenoms[0], reservePoolDenoms[1])

		orderAmount := reserveAmtY.Mul(sdk.MinDec(priceDiff.Quo(sdk.NewDec(2)).Abs(), sdk.NewDecWithPrec(1, 2))) // LUNA = LUNARESERVE * MIN(abs(PRICEDIFF/2),0.01)
		poolCreator := accAddr
		poolId := uint64(2) // TODO: query pool id for generalization
		swapTypeId := uint32(1)
		offerCoin := sdk.NewCoin(cfg.FireStation.DenomA, orderAmount.RoundInt()) // truncated
		demandCoinDenom := cfg.FireStation.DenomB
		orderPrice := globalPrice
		swapFeeRate := sdk.NewDecWithPrec(3, 3)

		msg, err := tx.MsgSwap(poolCreator, poolId, swapTypeId, offerCoin, demandCoinDenom, orderPrice, swapFeeRate)
		if err != nil {
			return fmt.Errorf("failed to create swap message: %s", err)
		}

		txBytes, err := transaction.Sign(ctx, accSeq, accNum, privKey, msg)
		if err != nil {
			return fmt.Errorf("failed to sign swap message: %s", err)
		}

		resp, err := transaction.BroadcastTx(ctx, txBytes)
		if err != nil {
			return fmt.Errorf("failed to broadcast transaction: %s", err)
		}

		log.Debug().
			Str("poolCreator", poolCreator).
			Uint64("poolId", poolId).
			Uint32("swapTypeId", swapTypeId).
			Str("offerCoin", offerCoin.String()).
			Str("demandCoinDenom", demandCoinDenom).
			Str("orderPrice", orderPrice.String()).
			Str("swapFeeRate", swapFeeRate.String()).
			Msg("")

		log.Info().
			Str("TxHash", resp.GetTxResponse().TxHash).
			Int64("Height", resp.GetTxResponse().Height).
			Msg("result")

	// LUNA is underpriced / ATOM is overpriced / price diff is less than -10%
	case priceDiff.IsNegative() && priceDiff.LT(sdk.NewDecWithPrec(-1, 1)):
		log.Info().Msgf("priceDiff is negative; selling '%s' and buying '%s'", reservePoolDenoms[0], reservePoolDenoms[1])

		orderAmount := reserveAmtX.Mul(sdk.MinDec(priceDiff.Quo(sdk.NewDec(2)).Abs(), sdk.NewDecWithPrec(1, 2))) // ATOM = ATOMRESERVE * MIN(abs(PRICEDIFF/2),0.01)
		poolCreator := accAddr
		poolId := uint64(2) // TODO: query pool id for generalization
		swapTypeId := uint32(1)
		offerCoin := sdk.NewCoin(cfg.FireStation.DenomA, orderAmount.RoundInt()) // truncated
		demandCoinDenom := cfg.FireStation.DenomB
		orderPrice := globalPrice
		swapFeeRate := sdk.NewDecWithPrec(3, 3)

		msg, err := tx.MsgSwap(poolCreator, poolId, swapTypeId, offerCoin, demandCoinDenom, orderPrice, swapFeeRate)
		if err != nil {
			return fmt.Errorf("failed to create swap message: %s", err)
		}

		txBytes, err := transaction.Sign(ctx, accSeq, accNum, privKey, msg)
		if err != nil {
			return fmt.Errorf("failed to sign swap message: %s", err)
		}

		resp, err := transaction.BroadcastTx(ctx, txBytes)
		if err != nil {
			return fmt.Errorf("failed to broadcast transaction: %s", err)
		}

		log.Debug().
			Str("poolCreator", poolCreator).
			Uint64("poolId", poolId).
			Uint32("swapTypeId", swapTypeId).
			Str("offerCoin", offerCoin.String()).
			Str("demandCoinDenom", demandCoinDenom).
			Str("orderPrice", orderPrice.String()).
			Str("swapFeeRate", swapFeeRate.String()).
			Msg("")

		log.Info().
			Str("TxHash", resp.GetTxResponse().TxHash).
			Int64("Height", resp.GetTxResponse().Height).
			Msg("result")
	default:
		log.Info().Msg("pool price is already stabilized")
	}

	return nil
}
