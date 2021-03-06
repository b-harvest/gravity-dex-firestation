package market

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/b-harvest/gravity-dex-firestation/config"

	sdk "github.com/cosmos/cosmos-sdk/types"

	resty "github.com/go-resty/resty/v2"
)

const (
	backendBaseAPIURL = "https://competition.bharvest.io:8081/"

	// [Deprecated]
	cmcAPIBaseURL       = "https://pro-api.coinmarketcap.com/"
	cmcAPIKeyHeaderName = "X-CMC_PRO_API_KEY"
	currency            = "USD"
)

type Client struct {
	client *resty.Client
	cfg    config.CoinMarketCapConfig
}

// NewClient creates new resty client.
func NewClient(cfg config.CoinMarketCapConfig) *Client {
	client := resty.New().SetHostURL(cmcAPIBaseURL).SetTimeout(time.Duration(5 * time.Second))
	return &Client{
		client: client,
		cfg:    cfg,
	}
}

func (c *Client) GetGlobalPrices(ctx context.Context, targetDenoms []string) ([]sdk.Dec, error) {
	client := resty.New().SetHostURL(backendBaseAPIURL).SetTimeout(time.Duration(5 * time.Second))

	resp, err := client.R().Get("prices")
	if err != nil {
		return []sdk.Dec{}, err
	}

	if resp.IsError() {
		return []sdk.Dec{}, err
	}

	type PricesData struct {
		BlockHeight int64              `json:"blockHeight"`
		Prices      map[string]float64 `json:"prices"`
		UpdatedAt   time.Time          `json:"updatedAt"`
	}

	var data PricesData
	err = json.Unmarshal(resp.Body(), &data)
	if err != nil {
		return []sdk.Dec{}, err
	}

	var result []sdk.Dec

	for _, d := range targetDenoms {
		denom := data.Prices[d[1:]]
		result = append(result, sdk.MustNewDecFromStr(strconv.FormatFloat(denom, 'f', 6, 64)))
	}

	return result, nil
}

type PoolsCache struct {
	BlockHeight      int64            `json:"blockHeight"`
	Pools            []PoolsCachePool `json:"pools"`
	TotalValueLocked float64          `json:"totalValueLocked"`
	UpdatedAt        time.Time        `json:"updatedAt"`
}
type PoolsCachePool struct {
	ID                        uint64           `json:"id"`
	ReserveCoins              []PoolsCacheCoin `json:"reserveCoins"`
	PoolCoin                  PoolsCacheCoin   `json:"poolCoin"`
	SwapFeeValueSinceLastHour float64          `json:"swapFeeValueSinceLastHour"`
	APY                       float64          `json:"apy"`
}
type PoolsCacheCoin struct {
	Denom       string  `json:"denom"`
	Amount      int64   `json:"amount"`
	GlobalPrice float64 `json:"globalPrice"`
}

func (c *Client) GetTargetPools(ctx context.Context) ([]uint64, error) {
	client := resty.New().SetHostURL(backendBaseAPIURL).SetTimeout(time.Duration(5 * time.Second))
	resp, err := client.R().Get("pools")
	if err != nil {
		return []uint64{}, err
	}
	if resp.IsError() {
		return []uint64{}, err
	}
	var data PoolsCache
	err = json.Unmarshal(resp.Body(), &data)
	if err != nil {
		return []uint64{}, err
	}

	var result []uint64

	for {
		var found bool

		rand.Seed(time.Now().UnixNano())
		randomIndex := rand.Intn(len(data.Pools))
		pool := data.Pools[randomIndex]

		if float64(pool.ReserveCoins[0].Amount)*pool.ReserveCoins[0].GlobalPrice > 1000000 &&
			float64(pool.ReserveCoins[1].Amount)*pool.ReserveCoins[1].GlobalPrice > 1000000 {

			for _, r := range result {
				if r == pool.ID {
					found = true
				}
			}

			if !found {
				result = append(result, pool.ID)
			}
		}

		if len(result) >= 4 {
			break
		}

		found = false
	}
	if len(result) == 4 {
		return result, err
	} else {
		return []uint64{}, fmt.Errorf("getTargetPools len")
	}
}

////////////////////////////////////////////////////////////////
// CoinMarketCap API

type CoinMarketCapResponse struct {
	Status struct {
		Timestamp    time.Time `json:"timestamp"`
		ErrorCode    int       `json:"error_code"`
		ErrorMessage string    `json:"error_message"`
		Elapsed      int       `json:"elapsed"`
		CreditCount  int       `json:"credit_count"`
	} `json:"status"`
	Data json.RawMessage `json:"data"`
}

// [Deprecated]
func (c *Client) request(ctx context.Context, params string, ids []string) (CoinMarketCapResponse, error) {
	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"id":      strings.Join(ids, ","),
			"convert": currency,
		}).
		SetHeader(cmcAPIKeyHeaderName, c.cfg.APIKey).
		Get(params)

	if err != nil {
		return CoinMarketCapResponse{}, err
	}

	if resp.IsError() {
		return CoinMarketCapResponse{}, err
	}

	var r CoinMarketCapResponse
	err = json.Unmarshal(resp.Body(), &r)
	if err != nil {
		return CoinMarketCapResponse{}, err
	}

	return r, nil
}
