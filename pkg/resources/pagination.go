package resources

const DefaultPageLimit = 20

type Page[T any] struct {
	Items []T `json:"items"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

func Paginate[T any](items []T, page, limit int) Page[T] {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = DefaultPageLimit
	}

	total := len(items)
	start := min((page-1)*limit, total)
	end := min(start+limit, total)

	return Page[T]{
		Items: items[start:end],
		Page:  page,
		Limit: limit,
		Total: total,
	}
}

func Filter[T any](items []T, keep func(T) bool) []T {
	out := make([]T, 0, len(items))
	for _, item := range items {
		if keep(item) {
			out = append(out, item)
		}
	}
	return out
}
