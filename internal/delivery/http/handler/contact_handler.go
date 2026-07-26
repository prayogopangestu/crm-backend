package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/response"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/request"
	contactusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/contact"
)

type ContactHandler struct {
	service *contactusecase.Service
	logger  *slog.Logger
}

func NewContactHandler(service *contactusecase.Service, logger *slog.Logger) *ContactHandler {
	return &ContactHandler{service: service, logger: logger}
}

func (h *ContactHandler) Routes(router chi.Router) {
	router.Get("/api/contacts", h.list)
	router.Post("/api/contacts", h.create)
	router.Put("/api/contacts/{id}", h.update)
	router.Delete("/api/contacts/{id}", h.delete)
}

func (h *ContactHandler) list(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.List(
		r.Context(), response.Principal(r), r.URL.Query().Get("search"), r.URL.Query().Get("status"),
		response.ParseInt(r.URL.Query().Get("page"), 1), response.ParseInt(r.URL.Query().Get("limit"), 20),
	)
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *ContactHandler) create(w http.ResponseWriter, r *http.Request) {
	var req request.ContactCreate
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

func (h *ContactHandler) update(w http.ResponseWriter, r *http.Request) {
	var req request.ContactCreate
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

func (h *ContactHandler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.Context(), response.Principal(r), chi.URLParam(r, "id")); err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "message": "Kontak berhasil dihapus"})
}
