package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/response"
	analyticsusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/analytics"
)

type AnalyticsHandler struct {
	service *analyticsusecase.Service
	logger  *slog.Logger
}

func NewAnalyticsHandler(service *analyticsusecase.Service, logger *slog.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{service: service, logger: logger}
}

func (h *AnalyticsHandler) Routes(router chi.Router) {
	router.Get("/api/dashboard/stats", h.dashboardStats)
	router.Get("/api/dashboard/conversion-chart", h.conversionChart)
	router.Get("/api/dashboard/activities", h.activities)
	router.Get("/api/reports/leaderboard", h.leaderboard)
	router.Get("/api/reports/lost-reasons", h.lostReasons)
	router.Get("/api/reports/goals", h.goals)
	router.Get("/api/reports/export/csv", h.exportCSV)
	router.Get("/api/reports/export/pdf", h.exportPDF)
}

func (h *AnalyticsHandler) dashboardStats(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.DashboardStats(r.Context(), response.Principal(r))
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *AnalyticsHandler) conversionChart(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.ConversionChart(r.Context(), response.Principal(r))
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *AnalyticsHandler) activities(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.Activities(r.Context(), response.Principal(r), response.ParseInt(r.URL.Query().Get("limit"), 5))
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *AnalyticsHandler) leaderboard(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.Leaderboard(r.Context(), response.Principal(r), r.URL.Query().Get("period"))
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *AnalyticsHandler) lostReasons(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.LostReasons(r.Context(), response.Principal(r))
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *AnalyticsHandler) goals(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.Goals(r.Context(), response.Principal(r))
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *AnalyticsHandler) exportCSV(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.ExportCSV(r.Context(), response.Principal(r))
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	writeDownload(w, "text/csv; charset=utf-8", "laporan-penjualan-"+time.Now().Format("2006-01-02")+".csv", data)
}

func (h *AnalyticsHandler) exportPDF(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.ExportPDF(r.Context(), response.Principal(r))
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	writeDownload(w, "application/pdf", "laporan-penjualan-"+time.Now().Format("2006-01-02")+".pdf", data)
}

func writeDownload(w http.ResponseWriter, contentType, filename string, data []byte) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
