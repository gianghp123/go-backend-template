package response

import "math"

type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type PaginatedResult[T any] struct {
	Data []T   `json:"data"`
	Meta *Meta `json:"meta"`
}

type PaginatedResponse[T any] struct {
	Data []T   `json:"data"`
	Meta *Meta `json:"meta"`
}

func NewMeta(page, limit int, total int64) *Meta {
	return &Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	}
}
