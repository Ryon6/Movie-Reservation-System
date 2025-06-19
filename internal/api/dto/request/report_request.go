package request

import "time"

type GenerateSalesReportRequest struct {
	MovieID   uint      `json:"movie_id" form:"movie_id" binding:"omitempty"`
	CinemaID  uint      `json:"cinema_id" form:"cinema_id" binding:"omitempty"`
	StartDate time.Time `json:"start_date" form:"start_date" binding:"omitempty"`
	EndDate   time.Time `json:"end_date" form:"end_date" binding:"omitempty"`
}
