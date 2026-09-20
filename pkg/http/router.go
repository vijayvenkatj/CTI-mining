package http

import (
	"net/http"

	"github.com/vijayvenkatj/cti-miner/pkg/http/controllers"
)

type Store interface {
	controllers.EdgeStore
	controllers.PulseStore
	controllers.IndicatorStore
}

func NewRouter(store Store) http.Handler {
	edges := controllers.NewEdgeController(store)
	pulses := controllers.NewPulseController(store)
	indicators := controllers.NewIndicatorController(store)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /edges", edges.List)
	mux.HandleFunc("GET /pulses", pulses.List)
	mux.HandleFunc("GET /pulses/{id}", pulses.Get)
	mux.HandleFunc("GET /indicators", indicators.List)
	mux.HandleFunc("GET /indicators/{id}", indicators.Get)
	return mux
}
