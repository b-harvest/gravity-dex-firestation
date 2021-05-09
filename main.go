package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/b-harvest/gravity-dex-firestation/client"
	"github.com/b-harvest/gravity-dex-firestation/config"
	"github.com/b-harvest/gravity-dex-firestation/tx"
	"github.com/b-harvest/gravity-dex-firestation/util"
	"github.com/b-harvest/gravity-dex-firestation/wallet"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// total amount of dollars worth of reserve coins to generate trading volume per hour
	remainingAmountPerHour = int64(1_000_000_000)

	// sending amount of dollars worth of coins to send for each buy and sell tx
	sendAmount = int64(69444)

	// sleep 1 seconds for each frequency
	frequency = 3600

	// number of hours
	duration = 22
)

func main() {
	cfg, err := config.Read(config.DefaultConfigPath)
	if err != nil {
		log.Fatalf("failed to read config: %s", err)
	}

	client, err := client.NewClient(cfg.RPC.Address, cfg.GRPC.Address, cfg.CoinMarketCap)
	if err != nil {
		log.Fatalf("failed to create new config: %s", err)
	}

	for i := 0; i < duration; i++ {
		impactTradingVolume(cfg, client)

		time.Sleep(1 * time.Hour)
	}
}

func impactTradingVolume(cfg config.Config, client *client.Client) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	fees := sdk.NewCoins(sdk.NewCoin(cfg.FireStation.FeeDenom, sdk.NewInt(cfg.FireStation.FeeAmount)))
	transaction := tx.NewTransaction(client, chainID, fees)

	log.Println("----------------------------------------------------------------")
	log.Printf("| ✅ ChainID: %s\n", chainID)
	log.Printf("| ✅ Sender: %s\n", accAddr)
	log.Printf("| ✅ Fees: %s\n", fees.String())

	pools, _ := client.GRPC.GetAllPools(context.Background())
	pools = util.Shuffle(pools)   // shuffle the exisiting pools and remove the ones that are not listed in CoinMarketCap
	pools = util.Select(pools, 4) // select n number of pools

	log.Println("----------------------------------------------------------------[Random Pools]")
	log.Printf("| pool 1: %s\n", pools[0].String())
	log.Printf("| pool 2: %s\n", pools[1].String())
	log.Printf("| pool 3: %s\n", pools[2].String())
	log.Printf("| pool 4: %s\n", pools[3].String())
	log.Printf("| pool 1 ReserveCoinDenoms: %s\n", pools[0].ReserveCoinDenoms)
	log.Printf("| pool 2 ReserveCoinDenoms: %s\n", pools[1].ReserveCoinDenoms)
	log.Printf("| pool 3 ReserveCoinDenoms: %s\n", pools[2].ReserveCoinDenoms)
	log.Printf("| pool 4 ReserveCoinDenoms: %s\n", pools[3].ReserveCoinDenoms)
	log.Println("----------------------------------------------------------------")

	targetDenoms := []string{
		pools[0].ReserveCoinDenoms[0],
		pools[0].ReserveCoinDenoms[1],
		pools[1].ReserveCoinDenoms[0],
		pools[1].ReserveCoinDenoms[1],
		pools[2].ReserveCoinDenoms[0],
		pools[2].ReserveCoinDenoms[1],
		pools[3].ReserveCoinDenoms[0],
		pools[3].ReserveCoinDenoms[1],
	}

	// request global prices only once to prevent from overuse
	globalPrices, err := client.Market.GetGlobalPrices(ctx, targetDenoms)
	if err != nil {
		return fmt.Errorf("failed to get pool prices: %s", err)
	}

	for i := 0; i < frequency; i++ {
		log.Printf("🔥 Trading Volume Bot🔥 %d out of %d frequency", i+1, frequency)

		var txBytes [][]byte

		for j, p := range pools {
			denomX := p.ReserveCoinDenoms[0]
			denomY := p.ReserveCoinDenoms[1]

			globalPriceX := globalPrices[2*j]
			globalPriceY := globalPrices[2*j+1]

			reserveAmtX, reserveAmtY, err := client.GRPC.GetPoolReserves(ctx, []string{denomX, denomY})
			if err != nil {
				return fmt.Errorf("failed to get pool price: %s", err)
			}

			reservePoolPrice := reserveAmtX.Quo(reserveAmtY)
			globalPrice := globalPriceY.Quo(globalPriceX)
			priceDiff := globalPrice.Quo(reservePoolPrice).Sub(sdk.NewDec(1))

			log.Println("----------------------------------------------------------------")
			log.Printf("| denomX: %s globalPriceX: %s\n", denomX, globalPriceX.String())
			log.Printf("| denomY: %s globalPriceY: %s\n", denomY, globalPriceY.String())

			poolCreator := accAddr
			poolId := p.GetPoolId()
			swapTypeId := uint32(1)
			swapFeeRate := sdk.NewDecWithPrec(3, 3)

			// swap denomY for denomX (buy)
			orderAmountX := sdk.NewDec(sendAmount / 4).Quo(globalPriceX).Mul(sdk.NewDec(1_000_000))
			offerCoinX := sdk.NewCoin(denomX, orderAmountX.RoundInt())         // truncated
			demandCoinDenomX := denomY                                         // the other side of pair
			orderPriceX := reservePoolPrice.Mul(sdk.MustNewDecFromStr("1.05")) // multiply pool price by 1.2 to buy higher price

			// swap denomX for denomY (sell)
			orderAmountY := sdk.NewDec(sendAmount / 4).Quo(globalPriceY).Mul(sdk.NewDec(1_000_000))
			offerCoinY := sdk.NewCoin(denomY, orderAmountY.RoundInt())         // truncated
			demandCoinDenomY := denomX                                         // the other side of pair
			orderPriceY := reservePoolPrice.Mul(sdk.MustNewDecFromStr("0.95")) // multiply pool price by 0.8 to sell cheaper price

			buyMsg, err := tx.MsgSwap(poolCreator, poolId, swapTypeId, offerCoinX, demandCoinDenomX, orderPriceX, swapFeeRate)
			if err != nil {
				return fmt.Errorf("failed to create swap message: %s", err)
			}

			buyMsg2, err := tx.MsgSwap(poolCreator, poolId, swapTypeId, offerCoinX, demandCoinDenomX, orderPriceX, swapFeeRate)
			if err != nil {
				return fmt.Errorf("failed to create swap message: %s", err)
			}

			sellMsg, err := tx.MsgSwap(poolCreator, poolId, swapTypeId, offerCoinY, demandCoinDenomY, orderPriceY, swapFeeRate)
			if err != nil {
				return fmt.Errorf("failed to create swap message: %s", err)
			}

			sellMsg2, err := tx.MsgSwap(poolCreator, poolId, swapTypeId, offerCoinY, demandCoinDenomY, orderPriceY, swapFeeRate)
			if err != nil {
				return fmt.Errorf("failed to create swap message: %s", err)
			}

			txByte, err := transaction.Sign(ctx, accSeq, accNum, privKey, buyMsg, buyMsg2, sellMsg, sellMsg2)
			if err != nil {
				return fmt.Errorf("failed to sign swap message: %s", err)
			}

			// increase sequence
			accSeq = accSeq + 1

			// decrease the remaining target amount of trading volume
			remainingAmountPerHour = remainingAmountPerHour - sendAmount

			txBytes = append(txBytes, txByte)

			log.Println("----------------------------------------------------------------[Common] [", j+1, " out of 4 pools]")
			log.Printf("| poolCreator: %s\n", poolCreator)
			log.Printf("| poolId: %d\n", poolId)
			log.Printf("| swapTypeId: %d\n", swapTypeId)
			log.Printf("| swapFeeRate: %s\n", swapFeeRate.String())
			log.Printf("| ✨ reservePoolPrice: %s\n", reservePoolPrice.String())
			log.Printf("| ✨ globalPrice: %s\n", globalPrice.String())
			log.Printf("| ✨ priceDiff : %s\n", priceDiff.String())
			log.Printf("| ✨ remainingAmountPerHour: %d\n", remainingAmountPerHour)
			log.Println("----------------------------------------------------------------[Swap Msg]")
			log.Printf("| ✅ globalPriceX: %s\n", globalPriceX.String())
			log.Printf("| ✅ orderAmountX: %s\n", orderAmountX.String())
			log.Printf("| ✅ offerCoinX: %s\n", offerCoinX.String())
			log.Printf("| ✅ demandCoinDenomX: %s\n", demandCoinDenomX)
			log.Printf("| ✅ orderPriceX: %s\n", orderPriceX)
			log.Println("----------------------------------------------------------------[Swap Msg]")
			log.Printf("| ✅ globalPriceY: %s\n", globalPriceY.String())
			log.Printf("| ✅ orderAmountY: %s\n", orderAmountY.String())
			log.Printf("| ✅ offerCoinY: %s\n", offerCoinY.String())
			log.Printf("| ✅ demandCoinDenomY: %s\n", demandCoinDenomY)
			log.Printf("| ✅ orderPriceY: %s\n", orderPriceY)
		}

		for k, txByte := range txBytes {
			resp, err := transaction.BroadcastTx(ctx, txByte)
			if err != nil {
				return fmt.Errorf("failed to broadcast transaction: %s", err)
			}
			log.Println("----------------------------------------------------------------[Sending Tx] [", k+1, " out of 4 pools]")
			log.Printf("| TxHash: %s\n", resp.GetTxResponse().TxHash)
			log.Printf("| Height: %d\n", resp.GetTxResponse().Height)
		}

		log.Println("----------------------------------------------------------------")
		fmt.Println("")
		fmt.Println("")

		time.Sleep(1 * time.Second)
	}

	return nil
}
