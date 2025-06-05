package cinema

import "mrs/internal/domain/shared/vo"

type SeatType string

const (
	SeatTypeStandard   = "STANDARD"
	SeatTypeVIP        = "VIP"
	SeatTypeWheelchair = "WHEELCHAIR"
)

// 注意：座位的可用性通常与特定的Showtime相关，而不是座位本身的静态属性。
type Seat struct {
	ID vo.SeatID
	// 对 CinemaHallID 单独建立索引: 应对高频查询影厅全部座位的需求
	CinemaHallID  vo.CinemaHallID
	RowIdentifier string // 座位所在排的标识,如 A、B、C
	SeatNumber    string // 座位在该排中的编号,如 1、2、3
	Type          SeatType
}
