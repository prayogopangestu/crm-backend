package websocket

import (
	"log/slog"
	"net/http"
)

type Hub struct {
	logger *slog.Logger
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{logger: logger}
}

func (h *Hub) Handle(w http.ResponseWriter, r *http.Request) {
	h.logger.Warn("websocket not yet implemented")
	http.Error(w, "websocket not implemented", http.StatusNotImplemented)
}
