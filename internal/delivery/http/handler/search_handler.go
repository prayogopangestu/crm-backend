package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/response"
	searchusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/search"
)

type SearchHandler struct {
	service *searchusecase.Service
	logger  *slog.Logger
}

func NewSearchHandler(service *searchusecase.Service, logger *slog.Logger) *SearchHandler {
	return &SearchHandler{service: service, logger: logger}
}

func (h *SearchHandler) Routes(router chi.Router) {
	router.Get("/api/search", h.search)
}

func (h *SearchHandler) search(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.Search(r.Context(), response.Principal(r), r.URL.Query().Get("q"))
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}
