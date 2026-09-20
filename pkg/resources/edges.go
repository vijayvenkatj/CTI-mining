package resources

import "fmt"

type Edge struct {
	ID     int64 `json:"id,omitempty"`
	Source string
	Target string
}

func MakeEdge(a, b string) Edge {
	if b > a {
		a, b = b, a
	}
	return Edge{
		Source: a,
		Target: b,
	}
}

func (e *Edge) String() string {
	return fmt.Sprintf("%s:%s", e.Source, e.Target)
}
