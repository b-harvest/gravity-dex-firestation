package market_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	resty "github.com/go-resty/resty/v2"
	"github.com/test-go/testify/require"
)

var (
	c                 *resty.Client
	backendBaseAPIURL = "https://competition.bharvest.io:8081/"
)

func TestMain(m *testing.M) {
	c = resty.New().SetHostURL(backendBaseAPIURL).SetTimeout(time.Duration(5 * time.Second))

	os.Exit(m.Run())
}

func TestParsePrices(t *testing.T) {
	resp, err := c.R().Get("prices")
	require.NoError(t, err)

	type PricesData struct {
		BlockHeight int64              `json:"blockHeight"`
		Prices      map[string]float64 `json:"prices"`
		UpdatedAt   time.Time          `json:"updatedAt"`
	}

	var data PricesData

	err = json.Unmarshal(resp.Body(), &data)
	require.NoError(t, err)

	fmt.Println("resp: ", data.BlockHeight)
	fmt.Println("resp: ", data.UpdatedAt)
	fmt.Println("resp: ", data.Prices["atom"])
}
