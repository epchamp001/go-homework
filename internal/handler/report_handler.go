package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"pvz-cli/internal/usecase"
)

type ReportsHandler interface {
	DownloadClientReport(c *gin.Context)
}

type reportsHandlerImp struct {
	svc usecase.Service
}

func NewReportsHandler(svc usecase.Service) *reportsHandlerImp {
	return &reportsHandlerImp{svc: svc}
}

func (h *reportsHandlerImp) DownloadClientReport(c *gin.Context) {
	sortBy := c.Query("sortBy")
	dataBytes, err := h.svc.GenerateClientReportByte(c, sortBy)
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to generate report: %v", err)
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=clients_report.xlsx")
	c.Header("Content-Length", fmt.Sprintf("%d", len(dataBytes)))

	c.Writer.Write(dataBytes)
}
