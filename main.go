package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/b-harvest/gravity-dex-firestation/client"
	"github.com/b-harvest/gravity-dex-firestation/config"
	"github.com/b-harvest/gravity-dex-firestation/tx"
	"github.com/b-harvest/gravity-dex-firestation/wallet"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Read(config.DefaultConfigPath)
	if err != nil {
		log.Fatalf("failed to read config: %s", err)
	}

	client, err := client.NewClient(cfg.RPC.Address, cfg.GRPC.Address, cfg.CoinMarketCap)
	if err != nil {
		log.Fatalf("failed to create new config: %s", err)
	}

	for {
		stablizePoolPrice(ctx, cfg, client)
	}
}

// stablizePoolPrice stablizes pool price of ATOM/LUNA pool.
// This provides users arbitrage opportunity for overpriced luna by managing the pool price.
func stablizePoolPrice(ctx context.Context, cfg config.Config, client *client.Client) error {
	log.Println("stablizing pool price...")

	chainID, err := client.RPC.GetNetworkChainID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain id: %s", err)
	}

	accAddr, privKey, err := wallet.RecoverAccountFromMnemonic(cfg.Wallet.Mnemonic, "")
	if err != nil {
		return fmt.Errorf("failed to retrieve account and private key from mnemonic: %s", err)
	}

	fees := sdk.NewCoins(sdk.NewCoin(cfg.FireStation.FeeDenom, sdk.NewInt(cfg.FireStation.FeeAmount)))
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
	globalPrice := globalPriceY.Quo(globalPriceX)              // GLOBALPRICE = LUNAUSD/ATOMUSD
	priceDiff := globalPrice.Quo(poolPrice).Sub(sdk.NewDec(1)) // PRICEDIFF   = GLOBALPRICE/POOLPRICE - 1

	log.Println("-----------------------------------------------------------------")
	log.Printf("| chainID: %s\n", chainID)
	log.Printf("| fees: %s\n", fees.String())
	log.Printf("| reserveAmtX: %s\n", reserveAmtX.String())
	log.Printf("| reserveAmtY: %s\n", reserveAmtY.String())
	log.Printf("| globalPriceX: %s\n", globalPriceX.String())
	log.Printf("| globalPriceY: %s\n", globalPriceY.String())
	log.Printf("| ✨ reservePoolPrice: %s\n", poolPrice.String())
	log.Printf("| ✨ globalPrice: %s\n", globalPrice.String())
	log.Printf("| ✨ priceDiff : %s\n", priceDiff.String())
	log.Printf("| ✨ priceDiff.Abs(): %s\n", priceDiff.Abs().String())
	log.Println("-----------------------------------------------------------------")

	transaction := tx.NewTransaction(client, chainID, fees)

	switch {
	// price diff is greater than 20%
	case priceDiff.GTE(sdk.NewDecWithPrec(2, 1)):
		log.Printf("🔥 priceDiff is positive; selling '%s' buying '%s'\n", reservePoolDenoms[0], reservePoolDenoms[1])

		orderAmount := reserveAmtX.Mul(sdk.MinDec(priceDiff.Quo(sdk.NewDec(2)).Abs(), sdk.NewDecWithPrec(1, 2))) // ATOM = ATOMRESERVE * MIN(abs(PRICEDIFF/2),0.01)
		offerCoin := sdk.NewCoin(cfg.FireStation.DenomA, orderAmount.RoundInt())                                 // truncated
		poolCreator := accAddr
		poolId := cfg.FireStation.PoolId
		swapTypeId := uint32(1)
		demandCoinDenom := cfg.FireStation.DenomB
		orderPrice := globalPrice
		swapFeeRate := sdk.NewDecWithPrec(3, 3)

		msg, err := tx.MsgSwap(poolCreator, poolId, swapTypeId, offerCoin, demandCoinDenom, orderPrice, swapFeeRate)
		if err != nil {
			return fmt.Errorf("failed to create swap message: %s", err)
		}

		for i := 1; i < 100000; i++ {
			account, err := client.GRPC.GetBaseAccountInfo(ctx, accAddr)
			if err != nil {
				return fmt.Errorf("failed to get account information: %s", err)
			}

			accSeq := account.GetSequence()
			accNum := account.GetAccountNumber()

			reserveAmtX, reserveAmtY, err := client.GRPC.GetPoolReserves(ctx, reservePoolDenoms)
			if err != nil {
				return fmt.Errorf("failed to get pool price: %s", err)
			}

			poolPrice := reserveAmtX.Quo(reserveAmtY)                  // POOLPRICE   = ATOMRESERVE/LUNARESERVE
			priceDiff := globalPrice.Quo(poolPrice).Sub(sdk.NewDec(1)) // PRICEDIFF   = GLOBALPRICE/POOLPRICE - 1

			log.Println("-------------------------------------------------------------send tx for [", i, "] times")
			log.Printf("| poolCreator: %s\n", poolCreator)
			log.Printf("| poolId: %d\n", poolId)
			log.Printf("| swapTypeId: %d\n", swapTypeId)
			log.Printf("| offerCoin: %s\n", offerCoin.String())
			log.Printf("| demandCoinDenom: %s\n", demandCoinDenom)
			log.Printf("| orderPrice: %s\n", orderPrice.String())
			log.Printf("| swapFeeRate: %s\n", swapFeeRate.String())
			log.Println("-------------------------------------------------------------")
			log.Printf("| reserveAmtX: %s\n", reserveAmtX.String())
			log.Printf("| reserveAmtY: %s\n", reserveAmtY.String())
			log.Printf("| globalPriceX: %s\n", globalPriceX.String())
			log.Printf("| globalPriceY: %s\n", globalPriceY.String())
			log.Printf("| ✨ reservePoolPrice: %s\n", poolPrice.String())
			log.Printf("| ✨ globalPrice: %s\n", globalPrice.String())
			log.Printf("| ✨ priceDiff: %s\n", priceDiff.String())
			log.Println("-------------------------------------------------------------")

			// exit when price diff is satified with the condition
			if priceDiff.Abs().LTE(sdk.NewDecWithPrec(1, 10)) {
				log.Println("gap between pool and global prices is 0.0000000001 percent now...!")
				os.Exit(1)
			}

			txBytes, err := transaction.Sign(ctx, accSeq, accNum, privKey, msg)
			if err != nil {
				return fmt.Errorf("failed to sign swap message: %s", err)
			}

			resp, err := transaction.BroadcastTx(ctx, txBytes)
			if err != nil {
				return fmt.Errorf("failed to broadcast transaction: %s", err)
			}

			log.Printf("TxHash: %s\n", resp.GetTxResponse().TxHash)
			log.Printf("Height: %d\n\n", resp.GetTxResponse().Height)

			time.Sleep(1 * time.Second)
		}

	// price diff is greater than -20%%
	case priceDiff.LTE(sdk.NewDecWithPrec(-2, 1)):
		log.Printf("🔥 priceDiff is negative; selling '%s' and buying '%s'\n", reservePoolDenoms[1], reservePoolDenoms[0])

		orderAmount := reserveAmtY.Mul(sdk.MinDec(priceDiff.Quo(sdk.NewDec(2)).Abs(), sdk.NewDecWithPrec(1, 2))) // LUNA = LUNARESERVE * MIN(abs(PRICEDIFF/2),0.01)
		offerCoin := sdk.NewCoin(cfg.FireStation.DenomB, orderAmount.RoundInt())                                 // truncated
		poolCreator := accAddr
		poolId := cfg.FireStation.PoolId
		swapTypeId := uint32(1)
		demandCoinDenom := cfg.FireStation.DenomA
		orderPrice := globalPrice
		swapFeeRate := sdk.NewDecWithPrec(3, 3)

		msg, err := tx.MsgSwap(poolCreator, poolId, swapTypeId, offerCoin, demandCoinDenom, orderPrice, swapFeeRate)
		if err != nil {
			return fmt.Errorf("failed to create swap message: %s", err)
		}

		for i := 1; i < 100000; i++ {
			account, err := client.GRPC.GetBaseAccountInfo(ctx, accAddr)
			if err != nil {
				return fmt.Errorf("failed to get account information: %s", err)
			}

			accSeq := account.GetSequence()
			accNum := account.GetAccountNumber()

			reserveAmtX, reserveAmtY, err := client.GRPC.GetPoolReserves(ctx, reservePoolDenoms)
			if err != nil {
				return fmt.Errorf("failed to get pool price: %s", err)
			}

			poolPrice := reserveAmtX.Quo(reserveAmtY)                  // POOLPRICE   = ATOMRESERVE/LUNARESERVE
			priceDiff := globalPrice.Quo(poolPrice).Sub(sdk.NewDec(1)) // PRICEDIFF   = GLOBALPRICE/POOLPRICE - 1

			log.Println("-------------------------------------------------------------send tx for [", i, "] times")
			log.Printf("| poolCreator: %s\n", poolCreator)
			log.Printf("| poolId: %d\n", poolId)
			log.Printf("| swapTypeId: %d\n", swapTypeId)
			log.Printf("| offerCoin: %s\n", offerCoin.String())
			log.Printf("| demandCoinDenom: %s\n", demandCoinDenom)
			log.Printf("| orderPrice: %s\n", orderPrice.String())
			log.Printf("| swapFeeRate: %s\n", swapFeeRate.String())
			log.Println("-------------------------------------------------------------")
			log.Printf("| reserveAmtX: %s\n", reserveAmtX.String())
			log.Printf("| reserveAmtY: %s\n", reserveAmtY.String())
			log.Printf("| globalPriceX: %s\n", globalPriceX.String())
			log.Printf("| globalPriceY: %s\n", globalPriceY.String())
			log.Printf("| ✨ reservePoolPrice: %s\n", poolPrice.String())
			log.Printf("| ✨ globalPrice: %s\n", globalPrice.String())
			log.Printf("| ✨ priceDiff: %s\n", priceDiff.String())
			log.Println("-------------------------------------------------------------")

			// exit when price diff is satified with the condition
			if priceDiff.Abs().LTE(sdk.NewDecWithPrec(1, 10)) {
				log.Println("❗ gap between pool and global prices is 0.0000000001 percent now ❗")
				os.Exit(1)
			}

			txBytes, err := transaction.Sign(ctx, accSeq, accNum, privKey, msg)
			if err != nil {
				return fmt.Errorf("failed to sign swap message: %s", err)
			}

			resp, err := transaction.BroadcastTx(ctx, txBytes)
			if err != nil {
				return fmt.Errorf("failed to broadcast transaction: %s", err)
			}

			log.Printf("TxHash: %s\n", resp.GetTxResponse().TxHash)
			log.Printf("Height: %d\n\n", resp.GetTxResponse().Height)

			time.Sleep(1 * time.Second)
		}

	default:
		log.Println("pool price is already stabilized")
	}

	return nil
}
