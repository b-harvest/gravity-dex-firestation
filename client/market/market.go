package market

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/b-harvest/gravity-dex-firestation/config"

	sdk "github.com/cosmos/cosmos-sdk/types"

	resty "github.com/go-resty/resty/v2"
)

const (
	backendBaseAPIURL = "https://competition.bharvest.io:8081/"

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

func (c *Client) GetMarketPrices(ctx context.Context, targetDenoms []string) ([]sdk.Dec, error) {
	prices, err := c.GetBackendPrices()
	if err != nil {
		return []sdk.Dec{}, fmt.Errorf("failed to get pool prices: %s", err)
	}

	var result []sdk.Dec

	for _, d := range targetDenoms {
		denom := prices[d[1:]]
		result = append(result, sdk.MustNewDecFromStr(strconv.FormatFloat(denom, 'f', 6, 64)))
	}

	return result, nil
}

type PricesData struct {
	BlockHeight int64              `json:"blockHeight"`
	Prices      map[string]float64 `json:"prices"`
	UpdatedAt   time.Time          `json:"updatedAt"`
}

func (c *Client) GetBackendPrices() (map[string]float64, error) {
	client := resty.New().SetHostURL(backendBaseAPIURL).SetTimeout(time.Duration(5 * time.Second))

	resp, err := client.R().Get("prices")
	if err != nil {
		return map[string]float64{}, err
	}

	if resp.IsError() {
		return map[string]float64{}, err
	}

	var data PricesData
	err = json.Unmarshal(resp.Body(), &data)
	if err != nil {
		return map[string]float64{}, err
	}

	return data.Prices, nil
}
