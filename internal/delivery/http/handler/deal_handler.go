package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/response"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/request"
	dealusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/deal"
)

type DealHandler struct {
	service *dealusecase.Service
	logger  *slog.Logger
}

func NewDealHandler(service *dealusecase.Service, logger *slog.Logger) *DealHandler {
	return &DealHandler{service: service, logger: logger}
}

func (h *DealHandler) Routes(router chi.Router) {
	router.Get("/api/deals", h.list)
	router.Post("/api/deals", h.create)
	router.Patch("/api/deals/{id}/stage", h.updateStage)
	router.Put("/api/deals/{id}", h.update)
	router.Delete("/api/deals/{id}", h.delete)
}

func (h *DealHandler) list(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.List(r.Context(), response.Principal(r))
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *DealHandler) create(w http.ResponseWriter, r *http.Request) {
	var req request.DealCreate
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

func (h *DealHandler) update(w http.ResponseWriter, r *http.Request) {
	var req request.DealCreate
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

func (h *DealHandler) updateStage(w http.ResponseWriter, r *http.Request) {
	var req request.DealStageUpdate
	if !response.DecodeJSON(w, r, &req) {
		return
	}
	if err := h.service.UpdateStage(r.Context(), response.Principal(r), chi.URLParam(r, "id"), req.ToInput()); err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "message": "Tahap deal berhasil diperbarui"})
}

func (h *DealHandler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.Context(), response.Principal(r), chi.URLParam(r, "id")); err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "message": "Deal berhasil dihapus"})
}
