package controllers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/vijayvenkatj/cti-miner/pkg/resources"
	"github.com/vijayvenkatj/cti-miner/pkg/storage"
)

type IndicatorStore interface {
	GetIndicator(ctx context.Context, id int64) (resources.Indicator, error)
}

type IndicatorController struct {
	store IndicatorStore
}

func NewIndicatorController(store IndicatorStore) *IndicatorController {
	return &IndicatorController{store: store}
}

func (c *IndicatorController) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid indicator id")
		return
	}

	indicator, err := c.store.GetIndicator(r.Context(), id)
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, "indicator not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, indicator)
}
