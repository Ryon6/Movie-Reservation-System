package response

type GenerateSalesReportResponse struct {
	// 报告基本信息
	ReportDate string `json:"report_date"` // 报告生成日期
	StartDate  string `json:"start_date"`  // 统计开始日期
	EndDate    string `json:"end_date"`    // 统计结束日期

	// 总体销售数据
	TotalRevenue  float64 `json:"total_revenue"`  // 总收入
	TotalBookings int     `json:"total_bookings"` // 总订单数
}
