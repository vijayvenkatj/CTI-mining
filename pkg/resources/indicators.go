package resources

import (
	"maps"
	"slices"
	"sync"
)

type Indicator struct {
	ID          int64   `json:"id"`
	Indicator   string  `json:"indicator"`
	Type        string  `json:"type"`
	Created     string  `json:"created"`
	Content     string  `json:"content"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Expiration  *string `json:"expiration"`
	IsActive    int     `json:"is_active"`
	Role        *string `json:"role"`

	PulseID   string
	PulseName string
}

// IndicatorIndex maps indicators that we get to their pulses.
type IndicatorIndex struct {
	m  map[string]map[string]struct{}
	mu sync.RWMutex
}

func (index *IndicatorIndex) Set(pulseID, indicatorID string) {
	index.mu.Lock()
	defer index.mu.Unlock()

	if index.m == nil {
		index.m = make(map[string]map[string]struct{})
	}
	indicatorMap := index.m[indicatorID]
	if indicatorMap == nil {
		indicatorMap = make(map[string]struct{})
		index.m[indicatorID] = indicatorMap
	}
	indicatorMap[pulseID] = struct{}{}
}

func (index *IndicatorIndex) Get(indicatorID string) []string {
	index.mu.RLock()
	defer index.mu.RUnlock()

	pulseMap := index.m[indicatorID]
	if pulseMap == nil {
		return []string{}
	}

	return slices.Sorted(maps.Keys(pulseMap))
}
