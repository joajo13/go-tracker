package sources_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joajo13/go-tracker/internal/domain"
	"github.com/joajo13/go-tracker/internal/sources"
)

func TestYahoo_FetchHappyPath(t *testing.T) {
	t.Parallel()

	fixture, err := os.ReadFile("testdata/yahoo_aapl.json")
	require.NoError(t, err)

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	defer srv.Close()

	adapter := sources.NewYahoo(sources.YahooConfig{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
		RatePerSec: 100,
		RateBurst:  100,
	})

	price, err := adapter.Fetch(context.Background(), "AAPL")

	require.NoError(t, err)
	assert.Equal(t, "/v8/finance/chart/AAPL", gotPath)
	assert.Equal(t, "yahoo", adapter.Name())
	assert.Equal(t, "172.45", price.Price.String())
	assert.Equal(t, domain.CurrencyUSD, price.Currency)
	assert.False(t, price.Ts.IsZero())
	assert.Equal(t, time.UTC, price.Ts.Location())
}

func TestYahoo_FetchHTTPError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	adapter := sources.NewYahoo(sources.YahooConfig{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
		RatePerSec: 100,
		RateBurst:  100,
	})

	_, err := adapter.Fetch(context.Background(), "AAPL")
	require.Error(t, err)
}

func TestYahoo_FetchEmptyResult(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"chart":{"result":[],"error":null}}`))
	}))
	defer srv.Close()

	adapter := sources.NewYahoo(sources.YahooConfig{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
		RatePerSec: 100,
		RateBurst:  100,
	})

	_, err := adapter.Fetch(context.Background(), "AAPL")
	require.Error(t, err)
}
