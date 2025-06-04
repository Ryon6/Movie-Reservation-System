package user

import "errors"

var (
	// 用户存在性错误
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")

	// 验证错误
	ErrInvalidUsername = errors.New("invalid username format")
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrWeakPassword    = errors.New("password does not meet strength requirements")
	ErrInvalidPassword = errors.New("invalid password")

	// 数据操作错误
	ErrDataConflict    = errors.New("data conflict")
	ErrVersionConflict = errors.New("version conflict")
)

var (
	// 存在性错误
	ErrRoleNotFound      = errors.New("role not found")
	ErrRoleAlreadyExists = errors.New("role already exists")

	// 操作错误
	ErrRoleInUse           = errors.New("role is in use by users")
	ErrDefaultRoleDeletion = errors.New("cannot delete default role")

	// 权限错误
	ErrRolePermissionDenied        = errors.New("role permission denied")
	ErrInvalidPermissionAssignment = errors.New("invalid permission assignment")

	// 验证错误
	ErrInvalidRoleName         = errors.New("invalid role name format")
	ErrInvalidPermissionFormat = errors.New("invalid permission format")
)
