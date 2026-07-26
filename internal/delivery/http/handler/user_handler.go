package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/response"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/request"
	userusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/user"
)

type UserHandler struct {
	service *userusecase.Service
	logger  *slog.Logger
}

func NewUserHandler(service *userusecase.Service, logger *slog.Logger) *UserHandler {
	return &UserHandler{service: service, logger: logger}
}

func (h *UserHandler) PublicRoutes(router chi.Router, middleware ...func(http.Handler) http.Handler) {
	router.With(middleware...).Post("/api/auth/login", h.login)
	router.With(middleware...).Post("/api/auth/register", h.register)
	router.With(middleware...).Post("/api/auth/accept-invite", h.acceptInvite)
}

func (h *UserHandler) ProtectedRoutes(router chi.Router) {
	router.Get("/api/profile", h.profile)
	router.Put("/api/profile", h.updateProfile)
	router.Get("/api/team", h.listTeam)
	router.Post("/api/team/invite", h.inviteMember)
	router.Delete("/api/team/{id}", h.revokeMember)
}

func (h *UserHandler) register(w http.ResponseWriter, r *http.Request) {
	var req request.UserRegister
	if !response.DecodeJSON(w, r, &req) {
		return
	}
	if _, err := h.service.Register(r.Context(), req.ToInput()); err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusCreated, map[string]any{"success": true, "message": "Registrasi berhasil"})
}

func (h *UserHandler) login(w http.ResponseWriter, r *http.Request) {
	var req request.UserLogin
	if !response.DecodeJSON(w, r, &req) {
		return
	}
	result, err := h.service.Login(r.Context(), req.ToInput())
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"token": result.Token, "user": result.User})
}

func (h *UserHandler) acceptInvite(w http.ResponseWriter, r *http.Request) {
	var req request.AcceptInvite
	if !response.DecodeJSON(w, r, &req) {
		return
	}
	if _, err := h.service.AcceptInvite(r.Context(), req.Token, req.Password); err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "message": "Undangan berhasil diaktifkan"})
}

func (h *UserHandler) profile(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.Profile(r.Context(), response.Principal(r))
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *UserHandler) updateProfile(w http.ResponseWriter, r *http.Request) {
	var req request.UpdateProfile
	if !response.DecodeJSON(w, r, &req) {
		return
	}
	result, err := h.service.UpdateProfile(r.Context(), response.Principal(r), req.ToInput())
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "message": "Profil diperbarui", "data": result})
}

func (h *UserHandler) listTeam(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.ListTeam(r.Context(), response.Principal(r))
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *UserHandler) inviteMember(w http.ResponseWriter, r *http.Request) {
	var req request.InviteMember
	if !response.DecodeJSON(w, r, &req) {
		return
	}
	result, err := h.service.InviteMember(r.Context(), response.Principal(r), req.ToInput())
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusCreated, map[string]any{"success": true, "data": result.User, "inviteUrl": result.InviteURL})
}

func (h *UserHandler) revokeMember(w http.ResponseWriter, r *http.Request) {
	if err := h.service.RevokeMember(r.Context(), response.Principal(r), chi.URLParam(r, "id")); err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "message": "Anggota tim dihapus"})
}
