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
	remainingAmountPerHour = int64(1_000_000_000) // total trading volume has to be $100,000,000 every hour.
	sendAmount             = int64(694_444)       // uses every frequency (694444 x 4 x 360 = 999999360 which is close to 1000000000)
	frequency              = 360
	duration               = 24
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

	// for i := 0; i < duration; i++ {
	// 	impactTradingVolume(cfg, client)

	// 	time.Sleep(1 * time.Hour)
	// }

	impactTradingVolume(cfg, client)
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
	// pools = util.Shuffle(pools)   // shuffle the exisiting pools and remove the ones that are not listed in CoinMarketCap
	pools = util.Random(pools, 4) // randomly pick 4 pools out of the existing pools

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

	// ids := []string{
	// 	types.CoinMarketCapMetadata[pools[0].ReserveCoinDenoms[0]],
	// 	types.CoinMarketCapMetadata[pools[0].ReserveCoinDenoms[1]],
	// 	types.CoinMarketCapMetadata[pools[1].ReserveCoinDenoms[0]],
	// 	types.CoinMarketCapMetadata[pools[1].ReserveCoinDenoms[1]],
	// 	types.CoinMarketCapMetadata[pools[2].ReserveCoinDenoms[0]],
	// 	types.CoinMarketCapMetadata[pools[2].ReserveCoinDenoms[1]],
	// 	types.CoinMarketCapMetadata[pools[3].ReserveCoinDenoms[0]],
	// 	types.CoinMarketCapMetadata[pools[3].ReserveCoinDenoms[1]],
	// }

	// // request global prices only once to prevent from overuse
	// market, err := client.Market.GetMarketPrices(ctx, ids)
	// if err != nil {
	// 	return fmt.Errorf("failed to get pool prices: %s", err)
	// }

	// ATOM := "27.89704229868265"
	// DVPN := ".02402116596556"
	// BTSG := "0.18456558759045"
	// XPRT := "9.30043142819725"
	// AKT := "5.17110114006091"

	// request global prices only once to prevent from overuse
	//globalPrices, err := client.Market.GetMarketPrices(ctx, ids)
	//if err != nil {
	//	return fmt.Errorf("failed to get pool prices: %s", err)
	//}

	// Localnet
	// 2021/05/10 01:56:57 | pool 1 ReserveCoinDenoms: [uatom ubtsg]
	// 2021/05/10 01:56:57 | pool 2 ReserveCoinDenoms: [uatom udvpn]
	// 2021/05/10 01:56:57 | pool 3 ReserveCoinDenoms: [uatom uxprt]
	// 2021/05/10 01:56:57 | pool 4 ReserveCoinDenoms: [uakt uatom]
	// globalPrices := []sdk.Dec{
	// 	sdk.MustNewDecFromStr("27.80916444485003"), sdk.MustNewDecFromStr("0.18455737143922"),
	// 	sdk.MustNewDecFromStr("27.80916444485003"), sdk.MustNewDecFromStr("0.02401638525683"),
	// 	sdk.MustNewDecFromStr("27.80916444485003"), sdk.MustNewDecFromStr("9.29970983876029"),
	// 	sdk.MustNewDecFromStr("5.18077974928129"), sdk.MustNewDecFromStr("27.80916444485003"),
	// }

	// Testnet
	// 2021/05/10 02:22:05 | pool 1 ReserveCoinDenoms: [uakt uatom]
	// 2021/05/10 02:22:05 | pool 2 ReserveCoinDenoms: [uatom uluna]
	// 2021/05/10 02:22:05 | pool 3 ReserveCoinDenoms: [uatom udvpn]
	// 2021/05/10 02:22:05 | pool 4 ReserveCoinDenoms: [uatom ubtsg]
	globalPrices := []sdk.Dec{
		sdk.MustNewDecFromStr("5.17940872752986"), sdk.MustNewDecFromStr("27.80916444485003"),
		sdk.MustNewDecFromStr("27.80916444485003"), sdk.MustNewDecFromStr("17.06741204554162"),
		sdk.MustNewDecFromStr("27.80916444485003"), sdk.MustNewDecFromStr("0.02422801943997"),
		sdk.MustNewDecFromStr("27.80916444485003"), sdk.MustNewDecFromStr("0.18818201792559"),
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

		time.Sleep(10 * time.Second)
	}

	return nil
}
