package controllers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/vijayvenkatj/cti-miner/pkg/resources"
)

type EdgeStore interface {
	ListEdges(ctx context.Context, afterID int64) ([]resources.Edge, error)
}

type EdgeController struct {
	store EdgeStore
}

func NewEdgeController(store EdgeStore) *EdgeController {
	return &EdgeController{store: store}
}

func (c *EdgeController) List(w http.ResponseWriter, r *http.Request) {
	afterID, _ := strconv.ParseInt(r.URL.Query().Get("after_id"), 10, 64)

	edges, err := c.store.ListEdges(r.Context(), afterID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if pulseID := r.URL.Query().Get("pulse_id"); pulseID != "" {
		edges = resources.Filter(edges, func(e resources.Edge) bool {
			return e.Source == pulseID || e.Target == pulseID
		})
	}

	page, limit := parsePage(r)
	writeJSON(w, http.StatusOK, resources.Paginate(edges, page, limit))
}
