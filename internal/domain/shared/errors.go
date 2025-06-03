package shared

import "errors"

// 通用错误
var (
	ErrInvalidInput   = errors.New("invalid input")
	ErrNoRowsAffected = errors.New("no rows affected")
)
