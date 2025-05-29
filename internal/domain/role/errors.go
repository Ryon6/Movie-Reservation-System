package role

import "errors"

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
