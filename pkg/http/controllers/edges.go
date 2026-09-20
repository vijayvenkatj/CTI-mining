package controllers

import (
	"context"
	"net/http"

	"github.com/vijayvenkatj/cti-miner/pkg/resources"
)

type EdgeStore interface {
	ListEdges(ctx context.Context) ([]resources.Edge, error)
}

type EdgeController struct {
	store EdgeStore
}

func NewEdgeController(store EdgeStore) *EdgeController {
	return &EdgeController{store: store}
}

func (c *EdgeController) List(w http.ResponseWriter, r *http.Request) {
	edges, err := c.store.ListEdges(r.Context())
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
