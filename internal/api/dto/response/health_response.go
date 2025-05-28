package response

// ComponentStatus 表示单个依赖组件的健康状态。
type ComponentStatus struct {
	Name    string `json:"name"`              // 组件名称
	Status  string `json:"status"`            // 组件状态，例如 "UP", "DOWN", "DEGRADED"
	Message string `json:"message,omitempty"` // 可选的额外信息或错误消息
}

type HealthResponse struct {
	OverallStatus string            `json:"overall_status"`       // 服务的总体健康状态
	Version       string            `json:"version,omitempty"`    // 应用程序版本 (可选)
	TimeStamp     string            `json:"timestamp"`            // 健康检查的时间戳
	Components    []ComponentStatus `json:"components,omitempty"` // 各依赖组件的状态 (可选)
}
