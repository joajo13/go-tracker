package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/time/rate"

	"github.com/joajo13/go-tracker/internal/domain"
)

// DolarAPIConfig configures a DolarAPI adapter.
type DolarAPIConfig struct {
	BaseURL    string
	HTTPClient *http.Client
	RatePerSec float64
	RateBurst  int
}

// DolarAPI is the PriceSource for FX rates (CCL/MEP/oficial/tarjeta) via
// dolarapi.com.
type DolarAPI struct {
	baseURL    string
	httpClient *http.Client
	limiter    *rate.Limiter
}

// NewDolarAPI builds the adapter with defaults filled in.
func NewDolarAPI(cfg DolarAPIConfig) *DolarAPI {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://dolarapi.com"
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 5 * time.Second}
	}
	if cfg.RatePerSec <= 0 {
		cfg.RatePerSec = 1
	}
	if cfg.RateBurst <= 0 {
		cfg.RateBurst = 2
	}
	return &DolarAPI{
		baseURL:    cfg.BaseURL,
		httpClient: cfg.HTTPClient,
		limiter:    rate.NewLimiter(rate.Limit(cfg.RatePerSec), cfg.RateBurst),
	}
}

// Name returns "dolarapi".
func (*DolarAPI) Name() string { return "dolarapi" }

// Fetch retrieves the latest sell price for the given dollar variant.
// Supported symbols: oficial, blue, contadoconliqui, mep, tarjeta.
func (d *DolarAPI) Fetch(ctx context.Context, symbol string) (domain.Price, error) {
	if err := d.limiter.Wait(ctx); err != nil {
		return domain.Price{}, fmt.Errorf("dolarapi rate wait: %w", err)
	}

	url := fmt.Sprintf("%s/v1/dolares/%s", d.baseURL, symbol)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return domain.Price{}, fmt.Errorf("dolarapi build request: %w", err)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return domain.Price{}, fmt.Errorf("dolarapi http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.Price{}, fmt.Errorf("dolarapi http %d", resp.StatusCode)
	}

	var payload struct {
		Casa               string  `json:"casa"`
		Venta              float64 `json:"venta"`
		FechaActualizacion string  `json:"fechaActualizacion"`
	}
	decodeErr := json.NewDecoder(resp.Body).Decode(&payload)
	if decodeErr != nil {
		return domain.Price{}, fmt.Errorf("dolarapi decode: %w", decodeErr)
	}

	ts, err := time.Parse(time.RFC3339, payload.FechaActualizacion)
	if err != nil {
		return domain.Price{}, fmt.Errorf("dolarapi ts %q: %w", payload.FechaActualizacion, err)
	}

	priceStr := strconv.FormatFloat(payload.Venta, 'f', -1, 64)

	return domain.NewPrice(&domain.PriceInput{
		TickerID: tickerIDPlaceholder(symbol),
		Source:   "dolarapi",
		Price:    priceStr,
		Currency: domain.CurrencyARS,
		Ts:       ts.UTC(),
	})
}
