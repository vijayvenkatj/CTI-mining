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
	for {
		msg, err := eg.Reader.ReadMessage(ctx)
		if err != nil {
			return err
		}

		var pulse resources.Pulse
		if err := json.Unmarshal(msg.Value, &pulse); err != nil {
			continue
		}

		for _, indicator := range pulse.Indicators {
			key := indicator.Indicator
			eg.CMS.Insert(key)

			// Check for very common indicators using CMS for freq
			if freq := eg.CMS.Estimate(key); freq >= eg.Threshold {
				eg.IndicatorIndex.Set(pulse.ID, key)
				continue
			}

			// Unique pulses -> make edges -> publish unique to writer
			pulses := eg.IndicatorIndex.Get(key)
			for _, target := range pulses {
				if target == pulse.ID {
					continue
				}

				edge := resources.MakeEdge(pulse.ID, target)
				if eg.Bloom.Contains(edge.String()) {
					continue
				}
				eg.Bloom.Insert(edge.String())

				value, err := json.Marshal(edge)
				if err != nil {
					return err
				}
				if err := eg.Writer.WriteMessage(ctx, []byte(edge.String()), value); err != nil {
					return err
				}
			}
			eg.IndicatorIndex.Set(pulse.ID, key)
		}
	}
}
