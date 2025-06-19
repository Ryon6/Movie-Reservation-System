package test

import (
	"fmt"
	"mrs/cmd/test/testutils"
	"mrs/internal/api/dto/request"
	"mrs/internal/api/dto/response"
	applog "mrs/pkg/log"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserManagementFlow(t *testing.T) {
	// 初始化测试服务器
	ts := testutils.NewTestServer(t)
	defer ts.Close()

	logger := ts.Logger.With(applog.String("Method", "TestUserManagementFlow"))

	// 1. 注册新用户
	registerReq := request.RegisterUserRequest{
		Username: "testuser",
		Password: "Test@123456",
		Email:    "testuser@example.com",
	}

	resp, _ := ts.DoRequest(t, http.MethodPost, "/api/v1/users/register", registerReq, "")
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// 2. 用户登录
	ts.UserToken = ts.Login(t, registerReq.Username, registerReq.Password)

	// 3. 获取用户个人信息
	resp, body := ts.DoRequest(t, http.MethodGet, "/api/v1/users/me", nil, ts.UserToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var profileResp response.UserProfileResponse
	testutils.ParseResponse(t, body, &profileResp)
	assert.Equal(t, registerReq.Username, profileResp.Username)
	assert.Equal(t, registerReq.Email, profileResp.Email)

	// 4. 管理员登录
	ts.AdminToken = ts.Login(t, "admin", "admin123")

	// 5. 创建新角色
	createRoleReq := request.CreateRoleRequest{
		Name:        "VIP会员",
		Description: "享有订票折扣特权的会员",
	}

	resp, body = ts.DoRequest(t, http.MethodPost, "/api/v1/admin/roles", createRoleReq, ts.AdminToken)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var roleResp response.RoleResponse
	testutils.ParseResponse(t, body, &roleResp)
	roleID := roleResp.ID

	// 6. 查询所有角色
	resp, body = ts.DoRequest(t, http.MethodGet, "/api/v1/admin/roles", nil, ts.AdminToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listRoleResp response.ListRoleResponse
	testutils.ParseResponse(t, body, &listRoleResp)
	assert.NotEmpty(t, listRoleResp.Roles)
	logger.Debug("\nlistRoleResp", applog.Any("listRoleResp", listRoleResp))

	// 7. 更新角色信息
	updateRoleReq := request.UpdateRoleRequest{
		ID:          roleID,
		Name:        "超级VIP会员",
		Description: "享有更多订票折扣特权的会员",
	}

	resp, body = ts.DoRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/admin/roles/%d", roleID), updateRoleReq, ts.AdminToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updatedRoleResp response.RoleResponse
	testutils.ParseResponse(t, body, &updatedRoleResp)
	assert.Equal(t, updateRoleReq.Name, updatedRoleResp.Name)
	assert.Equal(t, updateRoleReq.Description, updatedRoleResp.Description)

	// 8. 查询用户列表
	listUserReq := request.ListUserRequest{
		PaginationRequest: request.PaginationRequest{
			Page:     1,
			PageSize: 10,
		},
	}
	resp, body = ts.DoRequest(t, http.MethodGet, "/api/v1/admin/users", listUserReq, ts.AdminToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listUserResp response.ListUserResponse
	testutils.ParseResponse(t, body, &listUserResp)
	assert.NotEmpty(t, listUserResp.Users)

	// 找到新注册的用户
	var userID uint
	for _, u := range listUserResp.Users {
		if u.Username == registerReq.Username {
			userID = u.ID
			break
		}
	}
	assert.NotZero(t, userID, "应该能找到新注册的用户")

	// 9. 为用户分配角色
	assignRoleReq := request.AssignRoleToUserRequest{
		UserID: userID,
		RoleID: roleID,
	}

	resp, _ = ts.DoRequest(t, http.MethodPost, "/api/v1/admin/users/roles", assignRoleReq, ts.AdminToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 10. 验证用户角色已更新
	resp, body = ts.DoRequest(t, http.MethodGet, "/api/v1/users/me", nil, ts.UserToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updatedProfileResp response.UserProfileResponse
	testutils.ParseResponse(t, body, &updatedProfileResp)
	assert.Equal(t, updateRoleReq.Name, updatedProfileResp.RoleName)

	// 11. 删除角色（应该失败，因为有用户关联）
	resp, _ = ts.DoRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/admin/roles/%d", roleID), nil, ts.AdminToken)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// 12. 删除用户
	resp, _ = ts.DoRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/admin/users/%d", userID), nil, ts.AdminToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 13. 验证用户已被删除
	resp, _ = ts.DoRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/admin/users/%d", userID), nil, ts.AdminToken)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	// 14. 现在可以删除角色了
	resp, _ = ts.DoRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/admin/roles/%d", roleID), nil, ts.AdminToken)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
