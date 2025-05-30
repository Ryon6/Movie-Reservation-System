package user

import "errors"

var (
	// 用户存在性错误
	ErrUserExists = errors.New("user already exists")
	// ErrUsernameExists = errors.New("username already exists")
	// ErrEmailExists    = errors.New("email already exists")
	ErrUserNotFound = errors.New("user not found")

	// 验证错误
	ErrInvalidUsername = errors.New("invalid username format")
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrWeakPassword    = errors.New("password does not meet strength requirements")
	ErrInvalidPassword = errors.New("invalid password")

	// 数据操作错误
	ErrDataConflict    = errors.New("data conflict")
	ErrVersionConflict = errors.New("version conflict")
)
