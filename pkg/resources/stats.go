package resources

type Stats struct {
	TriangleEstimate int            `json:"triangle_estimate"`
	TotalPulses      int            `json:"total_pulses"`
	TotalIndicators  int            `json:"total_indicators"`
	TotalEdges       int            `json:"total_edges"`
	PulsesByTLP      map[string]int `json:"pulses_by_tlp"`
	IndicatorsByType map[string]int `json:"indicators_by_type"`
}
