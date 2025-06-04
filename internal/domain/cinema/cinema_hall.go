package cinema

import (
	"mrs/internal/domain/shared/vo"
)

// 影厅
type CinemaHall struct {
	ID          vo.CinemaHallID // 影厅ID
	Name        string          // 影厅名称
	ScreenType  string          // 屏幕类型
	SoundSystem string          // 音响系统
	Capacity    int             // 总座位数

	// 多对多关系
	Seats []*Seat // 聚合内部可以直接持有同一聚合内其他实体的引用
}
