// Package middleware contains HTTP middleware extracted from the application
// router: request id, recovery, access log, CORS, authentication and rate
// limiting. Each middleware is a self-contained constructor so the router can
// compose them without inline logic.
package middleware
