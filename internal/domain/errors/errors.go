// Package errors centralises the sentinel errors used across the domain,
// usecase and repository layers. Keeping them in a dedicated sub-package keeps
// the root domain package focused on value objects (Principal, Cache, etc.)
// and gives every caller a single, well-known place to import error sentinels
// from.
package errors

import "errors"

// Sentinel errors. Compare with errors.Is so callers can wrap them while still
// being matched by the HTTP error mapper in internal/delivery/http/httpx.
var (
	ErrNotFound       = errors.New("resource not found")
	ErrConflict       = errors.New("resource conflict")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrInvalidInput   = errors.New("invalid input")
	ErrStageInUse     = errors.New("pipeline stage is in use")
	ErrInviteExpired  = errors.New("invitation expired")
	ErrInviteUsed     = errors.New("invitation already used")
	ErrRedisRateLimit = errors.New("rate limit exceeded")
)
