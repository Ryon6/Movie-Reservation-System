package shared

import (
	"errors"
	"fmt"
)

// 通用错误
var (
	ErrInvalidInput   = errors.New("invalid input")
	ErrNoRowsAffected = errors.New("no rows affected")
)

var (
	// ErrTransactionBeginFailed 事务开启失败
	ErrTransactionBeginFailed = errors.New("failed to begin transaction")

	// ErrTransactionCommitFailed 事务提交失败
	ErrTransactionCommitFailed = errors.New("failed to commit transaction")

	// ErrTransactionRollbackFailed 事务回滚失败
	ErrTransactionRollbackFailed = errors.New("failed to rollback transaction")

	// ErrTransactionTimeout 事务超时
	ErrTransactionTimeout = errors.New("transaction timeout")
)

// TransactionError 表示事务相关的错误
type TransactionError struct {
	Op          string // 操作名称
	Err         error  // 原始错误
	RollbackErr error  // 回滚时的错误（如果有）
}

// Error 实现error接口
func (e *TransactionError) Error() string {
	if e.RollbackErr != nil {
		return fmt.Sprintf("%s: %v (rollback failed: %v)", e.Op, e.Err, e.RollbackErr)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// Unwrap 返回原始错误，支持errors.Is和errors.As
func (e *TransactionError) Unwrap() error {
	return e.Err
}

// NewTransactionError 创建新的事务错误
func NewTransactionError(op string, err error) *TransactionError {
	return &TransactionError{
		Op:  op,
		Err: err,
	}
}

// WithRollbackError 添加回滚错误信息
func (e *TransactionError) WithRollbackError(rollbackErr error) *TransactionError {
	e.RollbackErr = rollbackErr
	return e
}

// IsTransactionError 检查错误是否为事务相关错误
func IsTransactionError(err error) bool {
	var txErr *TransactionError
	return errors.As(err, &txErr)
}
