package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/response"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/request"
	integrationusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/integration"
)

type IntegrationHandler struct {
	service *integrationusecase.Service
	logger  *slog.Logger
}

func NewIntegrationHandler(service *integrationusecase.Service, logger *slog.Logger) *IntegrationHandler {
	return &IntegrationHandler{service: service, logger: logger}
}

func (h *IntegrationHandler) Routes(router chi.Router) {
	router.Get("/api/integrations/telegram", h.get)
	router.Put("/api/integrations/telegram", h.update)
	router.Post("/api/integrations/telegram/test", h.test)
}

func (h *IntegrationHandler) get(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.Get(r.Context(), response.Principal(r))
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *IntegrationHandler) update(w http.ResponseWriter, r *http.Request) {
	var req request.TelegramUpdate
	if !response.DecodeJSON(w, r, &req) {
		return
	}
	result, err := h.service.Update(r.Context(), response.Principal(r), req.ToInput())
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "enabled": result.Enabled})
}

func (h *IntegrationHandler) test(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Test(r.Context(), response.Principal(r)); err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "message": "Pesan uji coba terkirim"})
}
