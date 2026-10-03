package edgegen

import (
	"context"
	"encoding/json"

	"github.com/vijayvenkatj/cti-miner/pkg/algorithms"
	"github.com/vijayvenkatj/cti-miner/pkg/commons"
	"github.com/vijayvenkatj/cti-miner/pkg/resources"
)

type EdgeGenerator struct {
	Bloom *algorithms.BloomFilter
	CMS   *algorithms.CountMinSketch

	IndicatorIndex *resources.IndicatorIndex
	Threshold      uint64

	Reader *commons.KafkaReader
	Writer *commons.KafkaWriter
}

func NewEdgeGenerator(bloom *algorithms.BloomFilter, cms *algorithms.CountMinSketch, index *resources.IndicatorIndex, threshold uint64, reader *commons.KafkaReader, writer *commons.KafkaWriter) *EdgeGenerator {
	return &EdgeGenerator{
		Bloom:          bloom,
		CMS:            cms,
		IndicatorIndex: index,
		Threshold:      threshold,
		Reader:         reader,
		Writer:         writer,
	}
}

func (eg *EdgeGenerator) GenerateEdges(ctx context.Context) error {
	publish := func(edge resources.Edge) error {
		value, err := json.Marshal(edge)
		if err != nil {
			return err
		}
		return eg.Writer.WriteMessage(ctx, []byte(edge.String()), value)
	}

	for {
		msg, err := eg.Reader.ReadMessage(ctx)
		if err != nil {
			return err
		}

		var pulse resources.Pulse
		if err := json.Unmarshal(msg.Value, &pulse); err != nil {
			continue
		}

		if err := eg.ProcessPulse(pulse, publish); err != nil {
			return err
		}
	}
}

// ProcessPulse runs the CMS filter + indicator correlation for one pulse and
// calls emit for every new pulse-pulse edge. With a nil Bloom, dedup is
// skipped and every candidate edge is emitted (used by cmd/eval).
func (eg *EdgeGenerator) ProcessPulse(pulse resources.Pulse, emit func(resources.Edge) error) error {
	for _, indicator := range pulse.Indicators {
		key := indicator.Indicator
		eg.CMS.Insert(key)

		// Very common indicators make many weak edges: stop correlating on
		// them and free the pulse set they accumulated so far.
		if freq := eg.CMS.Estimate(key); freq >= eg.Threshold {
			eg.IndicatorIndex.Delete(key)
			continue
		}

		for _, target := range eg.IndicatorIndex.Get(key) {
			if target == pulse.ID {
				continue
			}

			edge := resources.MakeEdge(pulse.ID, target)
			if eg.Bloom != nil {
				if eg.Bloom.Contains(edge.String()) {
					continue
				}
				eg.Bloom.Insert(edge.String())
			}

			if err := emit(edge); err != nil {
				return err
			}
		}
		eg.IndicatorIndex.Set(pulse.ID, key)
	}
	return nil
}
