package request

// 创建订单请求
type CreateBookingRequest struct {
	UserID     uint
	ShowtimeID uint   `json:"showtime_id"`
	SeatIDs    []uint `json:"seat_ids"`
}
