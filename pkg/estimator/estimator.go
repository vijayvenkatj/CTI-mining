package estimator

import (
	"context"
	"encoding/json"
	"log"

	"github.com/vijayvenkatj/cti-miner/pkg/algorithms"
	"github.com/vijayvenkatj/cti-miner/pkg/commons"
	"github.com/vijayvenkatj/cti-miner/pkg/resources"
)

type EdgeStore interface {
	SaveEdge(ctx context.Context, edge resources.Edge) error
}

type Estimator struct {
	Triest *algorithms.Triest
	Reader *commons.KafkaReader
	Store  EdgeStore
}

func NewEstimator(triest *algorithms.Triest, reader *commons.KafkaReader) *Estimator {
	return &Estimator{
		Triest: triest,
		Reader: reader,
	}
}

func (e *Estimator) Run(ctx context.Context) error {
	for {
		msg, err := e.Reader.ReadMessage(ctx)
		if err != nil {
			return err
		}

		var edge resources.Edge
		if err := json.Unmarshal(msg.Value, &edge); err != nil {
			continue
		}

		e.Triest.Insert(edge)
		log.Println("triangles:", e.Triest.Estimate())

		if e.Store != nil {
			if err := e.Store.SaveEdge(ctx, edge); err != nil {
				log.Println("error storing edge", edge.String(), err)
			}
		}
	}
}
