package handlers

import (
	"mrs/internal/api/dto/request"
	"mrs/internal/app"
	applog "mrs/pkg/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportService app.ReportService
	logger        applog.Logger
}

func NewReportHandler(reportService app.ReportService, logger applog.Logger) *ReportHandler {
	return &ReportHandler{reportService: reportService, logger: logger.With(applog.String("Handler", "ReportHandler"))}
}

// GET /api/v1/admin/reports/sales 生成销售报告
func (h *ReportHandler) GenerateSalesReport(c *gin.Context) {
	var req request.GenerateSalesReportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("invalid request", applog.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.reportService.GenerateSalesReport(c, &req)
	if err != nil {
		h.logger.Error("generate sales report error", applog.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
	h.logger.Info("generate sales report successfully")
}
