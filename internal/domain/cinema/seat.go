package cinema

import (
	"fmt"
	"mrs/internal/domain/shared/vo"
)

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

func GenerateDefaultSeats(cinemaHallID vo.CinemaHallID) []*Seat {
	rows, columns := 10, 10
	seats := make([]*Seat, 0, rows*columns)
	for i := 0; i < rows; i++ {
		for j := 1; j <= columns; j++ {
			seats = append(seats, &Seat{
				CinemaHallID:  cinemaHallID,
				RowIdentifier: string(rune('A' + i)),
				SeatNumber:    fmt.Sprintf("%02d", j),
				Type:          SeatTypeStandard,
			})
		}
	}
	return seats
}

// 座位状态枚举
type SeatStatus int

const (
	SeatStatusAvailable SeatStatus = iota // 可用
	SeatStatusLocked                      // 已锁定
)

// 座位静态信息
type SeatInfo struct {
	ID            vo.SeatID  `json:"id"`             // 座位ID
	RowIdentifier string     `json:"row_identifier"` // 座位所在排的标识,如 A、B、C
	SeatNumber    string     `json:"seat_number"`    // 座位在该排中的编号,如 1、2、3
	Type          SeatType   `json:"type"`           // 座位类型
	Status        SeatStatus `json:"status"`         // 座位状态
}

// 获取座位显示名称
func (s *SeatInfo) GetDisplayName() string {
	return fmt.Sprintf("%s%s", s.RowIdentifier, s.SeatNumber)
}
