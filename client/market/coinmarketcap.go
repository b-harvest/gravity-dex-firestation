package market

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/b-harvest/gravity-dex-firestation/config"
	"github.com/b-harvest/gravity-dex-firestation/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	resty "github.com/go-resty/resty/v2"
)

const (
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

func (c *Client) GetMarketPrices(ctx context.Context, ids []string) ([]types.CoinMarketCapDataResult, error) {
	resp, err := c.request(ctx, "/v1/cryptocurrency/quotes/latest", ids)
	if err != nil {
		return []types.CoinMarketCapDataResult{}, fmt.Errorf("failed to get pool prices: %s", err)
	}

	var data map[string]struct {
		Id    int64 `json:"id"`
		Quote struct {
			USD struct {
				Price float64 `json:"price"`
			} `json:"USD"`
		} `json:"quote"`
	}

	err = json.Unmarshal(resp.Data, &data)
	if err != nil {
		return []types.CoinMarketCapDataResult{}, fmt.Errorf("failed to unmarshal market data: %s", err)
	}

	var result []types.CoinMarketCapDataResult
	for i, id := range ids {
		id = strings.ToUpper(id)

		d, ok := data[id]
		if !ok {
			return []types.CoinMarketCapDataResult{}, fmt.Errorf("price for the id %s not found", id)
		}

		if id == ids[i] {
			price, _ := sdk.NewDecFromStr(fmt.Sprintf("%f", d.Quote.USD.Price))

			temp := types.CoinMarketCapDataResult{
				Id:    id,
				Price: price,
			}
			result = append(result, temp)
		}
	}

	return result, nil
}
