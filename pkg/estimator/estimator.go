package estimator

import (
	"context"
	"encoding/json"

	"github.com/vijayvenkatj/cti-miner/pkg/algorithms"
	"github.com/vijayvenkatj/cti-miner/pkg/commons"
	"github.com/vijayvenkatj/cti-miner/pkg/resources"
)

type Estimator struct {
	Triest *algorithms.Triest
	Reader *commons.KafkaReader
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
	}
}
