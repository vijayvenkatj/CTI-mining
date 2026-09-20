package controllers

import (
	"context"
	"net/http"

	"github.com/vijayvenkatj/cti-miner/pkg/resources"
)

type StatsStore interface {
	GetStats(ctx context.Context) (resources.Stats, error)
}

type StatsController struct {
	store StatsStore
}

func NewStatsController(store StatsStore) *StatsController {
	return &StatsController{store: store}
}

func (c *StatsController) Get(w http.ResponseWriter, r *http.Request) {
	stats, err := c.store.GetStats(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stats)
}
