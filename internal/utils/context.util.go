package utils

import (
	"context"

	"your-project/internal/core/enums"
)

func GetCtx[T any](ctx context.Context, key enums.ContextKey) T {
	v, _ := ctx.Value(key).(T)
	return v
}
