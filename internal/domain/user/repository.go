package user

import (
	"context"
)

// UserRepository 定义了用户数据持久化操作的接口。
type UserRepository interface {
	Create(ctx context.Context, user *User) error                       // 创建用户
	FindByID(ctx context.Context, id uint) (*User, error)               // 通过ID获取用户
	FindByUsername(ctx context.Context, username string) (*User, error) // 通过name获取用户
	FindByEmail(ctx context.Context, email string) (*User, error)       // 通过email获取用户
	Update(ctx context.Context, user *User) error                       // 更新用户
	Delete(ctx context.Context, id uint) error                          // 删除用户
}

// type userRepository struct {
// 	db     *gorm.DB
// 	logger *zap.Logger
// }

// func NewUserRepository(db *gorm.DB, logger *zap.Logger) *userRepository {
// 	return &userRepository{
// 		db:     db,
// 		logger: logger,
// 	}
// }

// // 创建用户
// func (r *userRepository) CreateUser(user *User) error {
// 	Result := r.db.Create(&user)
// 	if Result.Error != nil {
// 		r.logger.Error("Failed to create user", zap.Error(Result.Error))
// 		return Result.Error
// 	}
// 	return nil
// }

// // 通过 ID 获取用户
// func (r *userRepository) GetUserById(id uint) (*User, error) {
// 	var user User
// 	Result := r.db.First(&user, id)
// 	if Result.Error != nil {
// 		r.logger.Error("Failed to get user by Id", zap.Uint("UserId", id), zap.Error(Result.Error))
// 		return nil, Result.Error
// 	}
// 	return &user, nil
// }

// // 通过用户名获取用户
// func (r *userRepository) GetUserByUsername(username string) (*User, error) {
// 	var user User
// 	Result := r.db.Where("username = ?", username).First(&user)
// 	if Result.Error != nil {
// 		r.logger.Error("Failed to get user by username", zap.String("username", username), zap.Error(Result.Error))
// 		return nil, Result.Error
// 	}
// 	return &user, nil
// }

// // 更新用户
// func (r *userRepository) UpdateUser(user *User) error {
// 	Result := r.db.Save(&user)
// 	if Result.Error != nil {
// 		r.logger.Error("Failed to update user", zap.Uint("UserId", user.ID), zap.Error(Result.Error))
// 		return Result.Error
// 	}
// 	return nil
// }

// // 删除用户
// func (r *userRepository) DeleteUser(id uint) error {
// 	Result := r.db.Delete(id)
// 	if Result.Error != nil {
// 		r.logger.Error("Failed to delete user", zap.Uint("UserId", id), zap.Error(Result.Error))
// 		return Result.Error
// 	}
// 	return nil
// }

// // 获取具有特定角色的用户
// func (r *userRepository) GetUsersByRole(roleName string) ([]*User, error) {
// 	var users []*User
// 	Result := r.db.Joins("JOIN roles ON roles.id = users.role_id").
// 		Where("roles.name = ?", roleName).
// 		Find(&users)
// 	if Result.Error != nil {
// 		r.logger.Error("Failed to get users by role", zap.String("roleName", roleName), zap.Error(Result.Error))
// 		return nil, Result.Error
// 	}
// 	return users, nil
// }
