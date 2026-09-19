package edgegen

import (
	"github.com/vijayvenkatj/cti-miner/pkg/algorithms"
	"github.com/vijayvenkatj/cti-miner/pkg/resources"
)

// TODO: replace chan with kafka reader and writer

type EdgeGenerator struct {
	Bloom *algorithms.BloomFilter
	CMS   *algorithms.CountMinSketch

	IndicatorIndex *resources.IndicatorIndex
	Threshold      uint64

	Writer chan resources.Edge
	Reader chan resources.Pulse
}

func NewEdgeGenerator(bloom *algorithms.BloomFilter, cms *algorithms.CountMinSketch, index *resources.IndicatorIndex, threshold uint64) *EdgeGenerator {
	return &EdgeGenerator{
		Bloom:          bloom,
		CMS:            cms,
		IndicatorIndex: index,
		Threshold:      threshold,
		Writer:         make(chan resources.Edge),
		Reader:         make(chan resources.Pulse),
	}
}

func (eg *EdgeGenerator) GenerateEdges() {
	for pulse := range eg.Reader {
		for _, indicator := range pulse.Indicators {
			key := indicator.Indicator
			eg.CMS.Insert(key)

			// Check for very common indicators using CMS for freq
			if freq := eg.CMS.Estimate(key); freq >= eg.Threshold {
				eg.IndicatorIndex.Set(pulse.ID, key)
				continue
			}

			// Unique pulses -> make edges -> send unique to writer
			pulses := eg.IndicatorIndex.Get(key)
			for _, target := range pulses {
				edge := resources.MakeEdge(pulse.ID, target)
				if eg.Bloom.Contains(edge.String()) {
					continue
				}
				eg.Bloom.Insert(edge.String())

				eg.Writer <- edge
			}
			eg.IndicatorIndex.Set(pulse.ID, key)
		}
	}
}
