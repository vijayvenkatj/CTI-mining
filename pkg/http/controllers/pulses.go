package controllers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/vijayvenkatj/cti-miner/pkg/resources"
	"github.com/vijayvenkatj/cti-miner/pkg/storage"
)

type PulseStore interface {
	GetPulse(ctx context.Context, id string) (resources.Pulse, error)
	ListPulses(ctx context.Context, modifiedSince string, ids []string) ([]resources.Pulse, error)
	PulseIndicators(ctx context.Context, pulseID string) ([]resources.Indicator, error)
}

type PulseController struct {
	store PulseStore
}

func NewPulseController(store PulseStore) *PulseController {
	return &PulseController{store: store}
}

func (c *PulseController) Get(w http.ResponseWriter, r *http.Request) {
	pulse, err := c.store.GetPulse(r.Context(), r.PathValue("id"))
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, "pulse not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, pulse)
}

func (c *PulseController) List(w http.ResponseWriter, r *http.Request) {
	var ids []string
	if raw := r.URL.Query().Get("ids"); raw != "" {
		ids = strings.Split(raw, ",")
	}
	modifiedSince := r.URL.Query().Get("modified_since")

	pulses, err := c.store.ListPulses(r.Context(), modifiedSince, ids)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	page, limit := parsePage(r)
	result := resources.Paginate(pulses, page, limit)

	if r.URL.Query().Get("include_indicators") == "true" {
		for i := range result.Items {
			indicators, err := c.store.PulseIndicators(r.Context(), result.Items[i].ID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			result.Items[i].Indicators = indicators
		}
	}

	writeJSON(w, http.StatusOK, result)
}
