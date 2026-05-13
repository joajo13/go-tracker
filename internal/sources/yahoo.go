package sources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/time/rate"

	"github.com/joajo13/go-tracker/internal/domain"
)

// YahooConfig configures a Yahoo adapter.
type YahooConfig struct {
	BaseURL    string
	HTTPClient *http.Client
	RatePerSec float64
	RateBurst  int
	UserAgent  string
}

// Yahoo is the PriceSource backed by Yahoo Finance's public chart endpoint.
type Yahoo struct {
	baseURL    string
	httpClient *http.Client
	limiter    *rate.Limiter
	userAgent  string
}

// NewYahoo builds a Yahoo adapter with defaults filled in.
func NewYahoo(cfg YahooConfig) *Yahoo {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://query1.finance.yahoo.com"
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 5 * time.Second}
	}
	if cfg.RatePerSec <= 0 {
		cfg.RatePerSec = 2
	}
	if cfg.RateBurst <= 0 {
		cfg.RateBurst = 4
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "go-tracker/0.1"
	}
	return &Yahoo{
		baseURL:    cfg.BaseURL,
		httpClient: cfg.HTTPClient,
		limiter:    rate.NewLimiter(rate.Limit(cfg.RatePerSec), cfg.RateBurst),
		userAgent:  cfg.UserAgent,
	}
}

// Name returns "yahoo".
func (*Yahoo) Name() string { return "yahoo" }

// Fetch retrieves the latest price for the given symbol.
func (y *Yahoo) Fetch(ctx context.Context, symbol string) (domain.Price, error) {
	if err := y.limiter.Wait(ctx); err != nil {
		return domain.Price{}, fmt.Errorf("yahoo rate wait: %w", err)
	}

	url := fmt.Sprintf("%s/v8/finance/chart/%s", y.baseURL, symbol)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return domain.Price{}, fmt.Errorf("yahoo build request: %w", err)
	}
	req.Header.Set("User-Agent", y.userAgent)

	resp, err := y.httpClient.Do(req)
	if err != nil {
		return domain.Price{}, fmt.Errorf("yahoo http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.Price{}, fmt.Errorf("yahoo http %d", resp.StatusCode)
	}

	var payload struct {
		Chart struct {
			Result []struct {
				Meta struct {
					Symbol             string  `json:"symbol"`
					RegularMarketPrice float64 `json:"regularMarketPrice"`
					Currency           string  `json:"currency"`
					RegularMarketTime  int64   `json:"regularMarketTime"`
				} `json:"meta"`
			} `json:"result"`
		} `json:"chart"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return domain.Price{}, fmt.Errorf("yahoo decode: %w", err)
	}
	if len(payload.Chart.Result) == 0 {
		return domain.Price{}, errors.New("yahoo: empty result")
	}

	meta := payload.Chart.Result[0].Meta
	currency := domain.Currency(meta.Currency)
	if !currency.IsValid() {
		return domain.Price{}, fmt.Errorf("yahoo: unsupported currency %q", meta.Currency)
	}

	priceStr := strconv.FormatFloat(meta.RegularMarketPrice, 'f', -1, 64)
	ts := time.Unix(meta.RegularMarketTime, 0).UTC()

	return domain.NewPrice(&domain.PriceInput{
		TickerID: tickerIDPlaceholder(symbol),
		Source:   "yahoo",
		Price:    priceStr,
		Currency: currency,
		Ts:       ts,
	})
}

// tickerIDPlaceholder lets us return a valid (non-zero) TickerID from the
// adapter even though the adapter does not know it. The caller (the worker)
// overwrites Price.TickerID with the real value before publishing.
func tickerIDPlaceholder(_ string) domain.TickerID { return 1 }
