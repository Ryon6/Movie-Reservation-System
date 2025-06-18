package handlers

import (
	"errors"
	"mrs/internal/api/dto/request"
	"mrs/internal/api/dto/response"
	"mrs/internal/api/middleware"
	"mrs/internal/app"
	"mrs/internal/domain/user"
	applog "mrs/pkg/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService app.UserService
	logger      applog.Logger
}

func NewUserHandler(userService app.UserService, logger applog.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger.With(applog.String("Handler", "UserHandler")),
	}
}

// 处理注册请求
func (h *UserHandler) Register(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "Register"))
	var req request.RegisterUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Warn("failed to bind register request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// 在service层判断DefaultRole是否为空
	userResp, err := h.userService.Register(ctx, &req)
	if err != nil {
		// 用户可能已存在
		if errors.Is(err, user.ErrUserAlreadyExists) {
			logger.Warn("User registration conflict", applog.Error(err))
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		logger.Error("user registration service failed", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("user registered successfully")
	ctx.JSON(http.StatusCreated, userResp)
}

// 用户获取自身信息
func (h *UserHandler) GetUserProfile(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "GetUserProfile"))
	id := ctx.GetUint(middleware.UserIDKey)

	userResp, err := h.userService.GetUser(ctx, &request.GetUserRequest{ID: id})
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			logger.Warn("user profile not found", applog.Uint("user_id", id))
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User profile not found"})
			return
		}
		logger.Error("Failed to get user profile", applog.Uint("user_id", id), applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error: User ID format issue"})
		return
	}

	logger.Info("user profile retrieved successfully", applog.Uint("user_id", id))
	ctx.JSON(http.StatusOK, userResp)
}

// 管理员获取用户信息
func (h *UserHandler) GetUser(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "GetUser"))
	id, err := getIDFromPath(ctx)
	if err != nil {
		logger.Error("failed to parse user_id", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	userResp, err := h.userService.GetUser(ctx, &request.GetUserRequest{ID: id})
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			logger.Warn("user not found", applog.Uint("user_id", uint(id)))
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		logger.Error("failed to get user", applog.Uint("user_id", uint(id)), applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	userProfileResp := response.UserProfileResponse{
		Username: userResp.Username,
		Email:    userResp.Email,
		RoleName: userResp.RoleName,
	}

	logger.Info("user retrieved successfully", applog.Uint("user_id", uint(id)))
	ctx.JSON(http.StatusOK, userProfileResp)
}

// 用户更新自身信息
func (h *UserHandler) UpdateUserProfile(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "UpdateUser"))
	var req request.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Warn("failed to bind update request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	id, err := getIDFromPath(ctx)
	if err != nil {
		logger.Error("failed to parse user_id", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	req.ID = uint(id)

	userResp, err := h.userService.UpdateUser(ctx, &req)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			logger.Warn("user not found", applog.Uint("user_id", id))
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		logger.Error("failed to update user", applog.Uint("user_id", id), applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	userProfileResp := response.UserProfileResponse{
		Username: userResp.Username,
		Email:    userResp.Email,
		RoleName: userResp.RoleName,
	}
	logger.Info("user profile updated successfully", applog.Uint("user_id", id))
	ctx.JSON(http.StatusOK, userProfileResp)
}

// 管理员更新用户信息
func (h *UserHandler) UpdateUser(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "AdminUpdateUser"))
	var req request.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Warn("failed to bind update request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	id, err := getIDFromPath(ctx)
	if err != nil {
		logger.Error("failed to parse user_id", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	req.ID = uint(id)

	userResp, err := h.userService.UpdateUser(ctx, &req)
	if err != nil {
		logger.Error("failed to update user", applog.Uint("user_id", uint(id)), applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("user profile updated successfully", applog.Uint("user_id", uint(id)))
	ctx.JSON(http.StatusOK, userResp)
}

// 删除用户
func (h *UserHandler) DeleteUser(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "DeleteUser"))
	id, err := getIDFromPath(ctx)
	if err != nil {
		logger.Error("failed to parse user_id", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.userService.DeleteUser(ctx, &request.DeleteUserRequest{ID: uint(id)})
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			logger.Warn("user not found", applog.Uint("user_id", uint(id)))
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		logger.Error("failed to delete user", applog.Uint("user_id", uint(id)), applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("user deleted successfully", applog.Uint("user_id", uint(id)))
	ctx.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// 获取用户列表
func (h *UserHandler) ListUsers(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "ListUsers"))
	var req request.ListUserRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		logger.Warn("failed to bind list users request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	userResp, err := h.userService.ListUsers(ctx, &req)
	if err != nil {
		logger.Error("failed to list users", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("list users successfully")
	ctx.JSON(http.StatusOK, userResp)
}

// 创建角色
func (h *UserHandler) CreateRole(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "CreateRole"))
	var req request.CreateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Warn("failed to bind create role request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	roleResp, err := h.userService.CreateRole(ctx, &req)
	if err != nil {
		if errors.Is(err, user.ErrRoleAlreadyExists) {
			logger.Warn("role already exists", applog.String("role_name", req.Name))
			ctx.JSON(http.StatusConflict, gin.H{"error": "Role already exists"})
			return
		}
		logger.Error("failed to create role", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("role created successfully")
	ctx.JSON(http.StatusCreated, roleResp)
}

// 获取角色列表
func (h *UserHandler) ListRoles(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "ListRoles"))
	roles, err := h.userService.ListRoles(ctx)
	if err != nil {
		logger.Error("failed to list roles", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("list roles successfully")
	ctx.JSON(http.StatusOK, roles)
}

// 更新角色
func (h *UserHandler) UpdateRole(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "UpdateRole"))

	id, err := getIDFromPath(ctx)
	if err != nil {
		logger.Error("failed to parse role_id", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var req request.UpdateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Warn("failed to bind update role request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}
	req.ID = id

	roleResp, err := h.userService.UpdateRole(ctx, &req)
	if err != nil {
		logger.Error("failed to update role", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("role updated successfully")
	ctx.JSON(http.StatusOK, roleResp)
}

// 删除角色
func (h *UserHandler) DeleteRole(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "DeleteRole"))
	id, err := getIDFromPath(ctx)
	if err != nil {
		logger.Error("failed to parse role_id", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	err = h.userService.DeleteRole(ctx, &request.DeleteRoleRequest{ID: id})
	if err != nil {
		if errors.Is(err, user.ErrRoleReferenced) {
			logger.Warn("role is referenced by users")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Role is referenced by users"})
			return
		}
		logger.Error("failed to delete role", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("role deleted successfully")
	ctx.JSON(http.StatusNoContent, nil)
}

// 为用户分配角色
func (h *UserHandler) AssignRoleToUser(ctx *gin.Context) {
	logger := h.logger.With(applog.String("Method", "AssignRoleToUser"))
	var req request.AssignRoleToUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Warn("failed to bind assign role to user request", applog.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	err := h.userService.AssignRoleToUser(ctx, &req)
	if err != nil {
		logger.Error("failed to assign role to user", applog.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("role assigned to user successfully")
	ctx.JSON(http.StatusOK, nil)
}
