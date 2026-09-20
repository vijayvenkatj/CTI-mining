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
	writeJSON(w, http.StatusOK, edges)
}
