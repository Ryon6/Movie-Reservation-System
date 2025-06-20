package request

type PaginationRequest struct {
	Page     int `json:"page" form:"page" binding:"required,min=1"`
	PageSize int `json:"page_size" form:"page_size" binding:"required,min=1,max=100"`
}
