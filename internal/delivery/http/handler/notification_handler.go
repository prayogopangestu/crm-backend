package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/response"
	notificationusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/notification"
)

type NotificationHandler struct {
	service *notificationusecase.Service
	logger  *slog.Logger
}

func NewNotificationHandler(service *notificationusecase.Service, logger *slog.Logger) *NotificationHandler {
	return &NotificationHandler{service: service, logger: logger}
}

func (h *NotificationHandler) Routes(router chi.Router) {
	router.Get("/api/notifications", h.list)
	router.Patch("/api/notifications/{id}/read", h.read)
	router.Patch("/api/notifications/read-all", h.readAll)
}

func (h *NotificationHandler) list(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.List(r.Context(), response.Principal(r))
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *NotificationHandler) read(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Read(r.Context(), response.Principal(r), chi.URLParam(r, "id")); err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true})
}

func (h *NotificationHandler) readAll(w http.ResponseWriter, r *http.Request) {
	if err := h.service.ReadAll(r.Context(), response.Principal(r)); err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true})
}
