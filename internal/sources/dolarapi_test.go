package sources_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joajo13/go-tracker/internal/domain"
	"github.com/joajo13/go-tracker/internal/sources"
)

func TestDolarApi_FetchCCL(t *testing.T) {
	t.Parallel()

	fixture, err := os.ReadFile("testdata/dolarapi_ccl.json")
	require.NoError(t, err)

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	defer srv.Close()

	adapter := sources.NewDolarAPI(sources.DolarAPIConfig{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
		RatePerSec: 100,
		RateBurst:  100,
	})

	price, err := adapter.Fetch(context.Background(), "contadoconliqui")

	require.NoError(t, err)
	assert.Equal(t, "/v1/dolares/contadoconliqui", gotPath)
	assert.Equal(t, "dolarapi", adapter.Name())
	assert.Equal(t, "1185.7", price.Price.String())
	assert.Equal(t, domain.CurrencyARS, price.Currency)
}

func TestDolarApi_FetchUnknownSymbol(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.NotFound(w, nil)
	}))
	defer srv.Close()

	adapter := sources.NewDolarAPI(sources.DolarAPIConfig{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
		RatePerSec: 100,
		RateBurst:  100,
	})

	_, err := adapter.Fetch(context.Background(), "ouija")
	require.Error(t, err)
}
