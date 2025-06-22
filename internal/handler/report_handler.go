package handler

import (
	"fmt"
	"net/http"
	"pvz-cli/internal/usecase/service"
	"pvz-cli/pkg/errs"

	"github.com/gin-gonic/gin"
)

type ReportsHandler interface {
	DownloadClientReport(c *gin.Context)
}

type reportsHandlerImp struct {
	svc service.Service
}

func NewReportsHandler(svc service.Service) *reportsHandlerImp {
	return &reportsHandlerImp{svc: svc}
}

func (h *reportsHandlerImp) DownloadClientReport(c *gin.Context) {
	sortBy := c.Query("sortBy")
	dataBytes, err := h.svc.GenerateClientReportByte(c, sortBy)
	if err != nil {
		msg := errs.ErrorCause(err)
		c.String(
			http.StatusInternalServerError,
			"failed to generate report: %s",
			msg,
		)
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=clients_report.xlsx")
	c.Header("Content-Length", fmt.Sprintf("%d", len(dataBytes)))

	c.Writer.Write(dataBytes)
}
