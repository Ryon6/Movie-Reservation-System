package request

// LoginRequest 定义了登录请求的结构体。
type LoginRequest struct {
	Username string `json:"username" binding:"required,alphanum,min=3,max=50"`
	Password string `json:"password" binding:"required,min=3,max=100"`
}
