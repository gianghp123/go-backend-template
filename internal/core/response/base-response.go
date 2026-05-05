package response

import "your-project/internal/core/errors"

type BaseResponse[T any] struct {
	Success bool            `json:"success"`
	Data    T               `json:"data,omitempty"`
	Error   *errors.AppError `json:"error,omitempty"`
	Meta    *Meta           `json:"meta,omitempty"`
}

func Success[T any](data T) BaseResponse[T] {
	return BaseResponse[T]{Success: true, Data: data}
}

func SuccessWithMeta[T any](data T, meta *Meta) BaseResponse[T] {
	return BaseResponse[T]{Success: true, Data: data, Meta: meta}
}

func Fail[T any](err *errors.AppError) BaseResponse[T] {
	return BaseResponse[T]{Success: false, Error: err}
}
