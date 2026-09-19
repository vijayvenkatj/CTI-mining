package resources

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
type IndicatorIndex map[string]map[string]struct{}