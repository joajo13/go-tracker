// Package api hosts the HTTP handlers exposed by the portfolio agent.
package api

import (
	"encoding/json"
	"net/http"
)

// HealthHandler returns a handler that reports the service is alive.
// Used by /healthz and any external monitoring (NFR-O-03).
func HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
}
