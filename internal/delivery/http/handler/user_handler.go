package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/response"
	"github.com/prayogopangestu/crm-system/backend/internal/delivery/http/request"
	"github.com/prayogopangestu/crm-system/backend/internal/domain"
	"github.com/prayogopangestu/crm-system/backend/internal/infrastructure/googleoauth"
	userusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/user"
)

type GoogleOAuthClient interface {
	AuthURL(state string) string
	Exchange(ctx context.Context, code string) (googleoauth.Profile, error)
}

type UserHandler struct {
	service *userusecase.Service
	logger  *slog.Logger
	google  GoogleOAuthClient
	baseURL string
	cache   domain.Cache
}

func NewUserHandler(service *userusecase.Service, logger *slog.Logger, google GoogleOAuthClient, baseURL string, cache domain.Cache) *UserHandler {
	return &UserHandler{service: service, logger: logger, google: google, baseURL: strings.TrimRight(baseURL, "/"), cache: cache}
}

func (h *UserHandler) PublicRoutes(router chi.Router, middleware ...func(http.Handler) http.Handler) {
	router.With(middleware...).Post("/api/auth/login", h.login)
	router.With(middleware...).Post("/api/auth/register", h.register)
	router.With(middleware...).Post("/api/auth/accept-invite", h.acceptInvite)
	router.Get("/api/auth/google/login", h.googleLogin)
	router.Get("/api/auth/google/callback", h.googleCallback)
}

func (h *UserHandler) ProtectedRoutes(router chi.Router) {
	router.Get("/api/profile", h.profile)
	router.Put("/api/profile", h.updateProfile)
	router.Get("/api/team", h.listTeam)
	router.Post("/api/team/invite", h.inviteMember)
	router.Delete("/api/team/{id}", h.revokeMember)
	router.Get("/api/workspaces", h.listWorkspaces)
	router.Post("/api/workspaces/switch", h.switchWorkspace)
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
	response.JSON(w, http.StatusOK, map[string]any{
		"token": result.Token, "user": result.User, "workspaces": result.Workspaces,
	})
}

func (h *UserHandler) acceptInvite(w http.ResponseWriter, r *http.Request) {
	var req request.AcceptInvite
	if !response.DecodeJSON(w, r, &req) {
		return
	}
	result, err := h.service.AcceptInvite(r.Context(), req.Token, req.Password)
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"token": result.Token, "user": result.User, "workspaces": result.Workspaces,
	})
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

func (h *UserHandler) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.ListWorkspaces(r.Context(), response.Principal(r))
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *UserHandler) switchWorkspace(w http.ResponseWriter, r *http.Request) {
	var req request.SwitchWorkspace
	if !response.DecodeJSON(w, r, &req) {
		return
	}
	result, err := h.service.SwitchWorkspace(r.Context(), response.Principal(r), req.ToInput())
	if err != nil {
		response.WriteError(h.logger, w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"token": result.Token, "workspace": result.Workspace,
	})
}

func (h *UserHandler) googleLogin(w http.ResponseWriter, r *http.Request) {
	if h.google == nil {
		response.Error(w, http.StatusServiceUnavailable, "google_disabled", "Login Google belum dikonfigurasi", response.RequestID(r.Context()), nil)
		return
	}
	state := randomState()
	if h.cache != nil {
		_ = h.cache.SetJSON(r.Context(), "crm:google:state:"+state, true, 10*time.Minute)
	}
	http.Redirect(w, r, h.google.AuthURL(state), http.StatusFound)
}

func (h *UserHandler) googleCallback(w http.ResponseWriter, r *http.Request) {
	if h.google == nil {
		redirectWithError(w, r, h.baseURL, "google_disabled")
		return
	}
	q := r.URL.Query()
	state := q.Get("state")
	code := q.Get("code")

	if state == "" || code == "" {
		redirectWithError(w, r, h.baseURL, "google_invalid_request")
		return
	}
	if h.cache != nil {
		var ok bool
		hit, err := h.cache.GetJSON(r.Context(), "crm:google:state:"+state, &ok)
		if err != nil || !hit {
			redirectWithError(w, r, h.baseURL, "google_state_invalid")
			return
		}
	}

	profile, err := h.google.Exchange(r.Context(), code)
	if err != nil {
		h.logger.Error("google exchange failed", "error", err)
		redirectWithError(w, r, h.baseURL, "google_exchange_failed")
		return
	}

	result, err := h.service.LoginWithGoogle(r.Context(), userusecase.GoogleProfileInput{
		GoogleID:  profile.ID,
		Email:     profile.Email,
		FirstName: firstNonEmpty(profile.GivenName, profile.Name),
		LastName:  profile.FamilyName,
		AvatarURL: profile.Picture,
	})
	if err != nil {
		redirectWithError(w, r, h.baseURL, "google_login_failed")
		return
	}

	target := h.baseURL + "/auth/google/callback?token=" + url.QueryEscape(result.Token)
	http.Redirect(w, r, target, http.StatusFound)
}

func redirectWithError(w http.ResponseWriter, r *http.Request, baseURL, code string) {
	target := baseURL + "/auth/google/callback?error=" + url.QueryEscape(code)
	http.Redirect(w, r, target, http.StatusFound)
}

func randomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
