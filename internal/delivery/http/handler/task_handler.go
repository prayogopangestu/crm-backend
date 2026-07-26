package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/response"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/request"
	taskusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/task"
)

type TaskHandler struct {
	service *taskusecase.Service
	logger  *slog.Logger
}

func NewTaskHandler(service *taskusecase.Service, logger *slog.Logger) *TaskHandler {
	return &TaskHandler{service: service, logger: logger}
}

func (h *TaskHandler) Routes(router chi.Router) {
	router.Get("/api/tasks", h.list)
	router.Post("/api/tasks", h.create)
	router.Put("/api/tasks/{id}", h.update)
	router.Patch("/api/tasks/{id}/toggle", h.toggle)
	router.Delete("/api/tasks/{id}", h.delete)
}

func (h *TaskHandler) list(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.List(r.Context(), response.Principal(r), r.URL.Query().Get("date"), r.URL.Query().Get("status"))
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *TaskHandler) create(w http.ResponseWriter, r *http.Request) {
	var req request.TaskCreate
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

func (h *TaskHandler) update(w http.ResponseWriter, r *http.Request) {
	var req request.TaskCreate
	if !response.DecodeJSON(w, r, &req) {
		return
	}
	result, err := h.service.Update(r.Context(), response.Principal(r), chi.URLParam(r, "id"), req.ToInput())
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "data": result})
}

func (h *TaskHandler) toggle(w http.ResponseWriter, r *http.Request) {
	var req request.TaskToggle
	if !response.DecodeJSON(w, r, &req) {
		return
	}
	if err := h.service.Toggle(r.Context(), response.Principal(r), chi.URLParam(r, "id"), req.Completed); err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "completed": req.Completed})
}

func (h *TaskHandler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.Context(), response.Principal(r), chi.URLParam(r, "id")); err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "message": "Tugas dihapus"})
}
