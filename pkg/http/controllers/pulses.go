package controllers

import (
	"context"
	"errors"
	"net/http"

	"github.com/vijayvenkatj/cti-miner/pkg/resources"
	"github.com/vijayvenkatj/cti-miner/pkg/storage"
)

type PulseStore interface {
	GetPulse(ctx context.Context, id string) (resources.Pulse, error)
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
