package http

import (
	"net/http"

	"github.com/vijayvenkatj/cti-miner/pkg/http/controllers"
)

type Store interface {
	controllers.EdgeStore
	controllers.PulseStore
	controllers.IndicatorStore
	controllers.StatsStore
}

func NewRouter(store Store) http.Handler {
	edges := controllers.NewEdgeController(store)
	pulses := controllers.NewPulseController(store)
	indicators := controllers.NewIndicatorController(store)
	stats := controllers.NewStatsController(store)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /edges", edges.List)
	mux.HandleFunc("GET /pulses", pulses.List)
	mux.HandleFunc("GET /pulses/{id}", pulses.Get)
	mux.HandleFunc("GET /indicators", indicators.List)
	mux.HandleFunc("GET /indicators/{id}", indicators.Get)
	mux.HandleFunc("GET /stats", stats.Get)
	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
