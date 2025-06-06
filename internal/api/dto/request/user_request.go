package request

// RegisterUserRequest 定义了用户注册请求的结构体。
type RegisterUserRequest struct {
	Username    string `json:"username" binding:"required,alphanum,min=3,max=50"`
	Password    string `json:"password" binding:"required,min=8,max=100"`
	Email       string `json:"email" binding:"required,email"`
	DefaultRole string `json:"default_role" binding:"omitempty,alphanum"` // 可选，如果允许客户端指定
}

type GetUserRequest struct {
	ID uint `json:"id" binding:"required,min=1"`
}
