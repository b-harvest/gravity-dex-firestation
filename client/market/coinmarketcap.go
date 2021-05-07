package market

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/b-harvest/gravity-dex-firestation/config"

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

func (c *Client) request(ctx context.Context, params string, symbols []string) (CoinMarketCapResponse, error) {
	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"symbol":  strings.Join(symbols, ","),
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

func (c *Client) GetMarketPrices(ctx context.Context, symbols []string) (sdk.Dec, sdk.Dec, error) {
	resp, err := c.request(ctx, "/v1/cryptocurrency/quotes/latest", symbols)
	if err != nil {
		return sdk.ZeroDec(), sdk.ZeroDec(), fmt.Errorf("failed to get pool prices: %s", err)
	}

	var data map[string]struct {
		Quote struct {
			USD struct {
				Price float64 `json:"price"`
			} `json:"USD"`
		} `json:"quote"`
	}

	err = json.Unmarshal(resp.Data, &data)
	if err != nil {
		return sdk.ZeroDec(), sdk.ZeroDec(), fmt.Errorf("failed to unmarshal market data: %s", err)
	}

	var priceX, priceY sdk.Dec
	for _, symbol := range symbols {
		symbol = strings.ToUpper(symbol)

		d, ok := data[symbol]
		if !ok {
			return sdk.ZeroDec(), sdk.ZeroDec(), fmt.Errorf("price for symbol %s not found", symbol)
		}

		// another way to convert float64 type to sdk.Dec
		// sdk.NewDecWithPrec(int64(f * 1000000), 6)
		if symbol == symbols[0] {
			priceX, _ = sdk.NewDecFromStr(fmt.Sprintf("%f", d.Quote.USD.Price))
		}

		if symbol == symbols[1] {
			priceY, _ = sdk.NewDecFromStr(fmt.Sprintf("%f", d.Quote.USD.Price))
		}
	}

	return priceX, priceY, nil
}
