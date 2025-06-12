package request

import (
	"mrs/internal/domain/shared/vo"
	"mrs/internal/domain/showtime"
	"time"
)

// 创建场次
type CreateShowtimeRequest struct {
	MovieID      uint      `json:"movie_id" binding:"required,min=1"`
	CinemaHallID uint      `json:"cinema_hall_id" binding:"required,min=1"`
	StartTime    time.Time `json:"start_time" binding:"required"`
	EndTime      time.Time `json:"end_time" binding:"required"`
	Price        float64   `json:"price" binding:"required,min=0"`
}

func (r *CreateShowtimeRequest) ToDomain() *showtime.Showtime {
	return &showtime.Showtime{
		MovieID:      vo.MovieID(r.MovieID),
		CinemaHallID: vo.CinemaHallID(r.CinemaHallID),
		StartTime:    r.StartTime,
		EndTime:      r.EndTime,
		Price:        r.Price,
	}
}

// 获取场次
type GetShowtimeRequest struct {
	ID uint
}

func (r *GetShowtimeRequest) ToDomain() *showtime.Showtime {
	return &showtime.Showtime{
		ID: vo.ShowtimeID(r.ID),
	}
}

// 更新场次
type UpdateShowtimeRequest struct {
	ID           uint
	MovieID      uint      `json:"movie_id" binding:"omitempty,min=1"`
	CinemaHallID uint      `json:"cinema_hall_id" binding:"omitempty,min=1"`
	StartTime    time.Time `json:"start_time" binding:"omitempty"`
	EndTime      time.Time `json:"end_time" binding:"omitempty"`
	Price        float64   `json:"price" binding:"omitempty,min=0"`
}

func (r *UpdateShowtimeRequest) ToDomain() *showtime.Showtime {
	return &showtime.Showtime{
		ID:           vo.ShowtimeID(r.ID),
		MovieID:      vo.MovieID(r.MovieID),
		CinemaHallID: vo.CinemaHallID(r.CinemaHallID),
		StartTime:    r.StartTime,
		EndTime:      r.EndTime,
		Price:        r.Price,
	}
}

// 删除场次
type DeleteShowtimeRequest struct {
	ID uint
}

func (r *DeleteShowtimeRequest) ToDomain() *showtime.Showtime {
	return &showtime.Showtime{
		ID: vo.ShowtimeID(r.ID),
	}
}

// 获取放映列表
type ListShowtimesRequest struct {
	PaginationRequest
	MovieID      uint      `json:"movie_id" binding:"omitempty,min=1"`
	CinemaHallID uint      `json:"cinema_hall_id" binding:"omitempty,min=1"`
	Date         time.Time `json:"date" binding:"omitempty"`
}

func (r *ListShowtimesRequest) ToDomain() *showtime.ShowtimeQueryOptions {
	return &showtime.ShowtimeQueryOptions{
		MovieID:      vo.MovieID(r.MovieID),
		CinemaHallID: vo.CinemaHallID(r.CinemaHallID),
		Date:         r.Date,
		Page:         r.Page,
		PageSize:     r.PageSize,
	}
}

// 获取指定放映场次的座位表（包含座位状态）
type GetSeatMapRequest struct {
	ShowtimeID uint
}
