package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/response"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/request"
	pipelineusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/pipeline"
)

type PipelineHandler struct {
	service *pipelineusecase.Service
	logger  *slog.Logger
}

func NewPipelineHandler(service *pipelineusecase.Service, logger *slog.Logger) *PipelineHandler {
	return &PipelineHandler{service: service, logger: logger}
}

func (h *PipelineHandler) Routes(router chi.Router) {
	router.Get("/api/pipeline-stages", h.list)
	router.Post("/api/pipeline-stages", h.create)
	router.Put("/api/pipeline-stages/reorder", h.reorder)
	router.Delete("/api/pipeline-stages/{id}", h.delete)
}

func (h *PipelineHandler) list(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.List(r.Context(), response.Principal(r))
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *PipelineHandler) create(w http.ResponseWriter, r *http.Request) {
	var req request.PipelineCreate
	if !response.DecodeJSON(w, r, &req) {
		return
	}
	result, err := h.service.Create(r.Context(), response.Principal(r), req.ToInput())
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusCreated, map[string]any{"success": true, "data": result})
}

func (h *PipelineHandler) reorder(w http.ResponseWriter, r *http.Request) {
	var req request.PipelineReorder
	if !response.DecodeJSON(w, r, &req) {
		return
	}
	if err := h.service.Reorder(r.Context(), response.Principal(r), req.StagesOrder); err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true})
}

func (h *PipelineHandler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.Context(), response.Principal(r), chi.URLParam(r, "id")); err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "message": "Tahapan dihapus"})
}
