package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var CoinMarketCapMetadata = map[string]string{
	"uatom": "3794",
	"ubtsg": "8905",
	"udvpn": "2643",
	"uxprt": "7281",
	"uakt":  "7431",
	"uluna": "4172",
	"ungm":  "8279",
	"uiris": "3874",
}

type CoinMarketCapDataResult struct {
	Id    string `json:"id"`
	Price sdk.Dec
}
