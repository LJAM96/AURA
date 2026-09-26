package middleware

import (
	"aura/config"
	"net/http"
	"slices"
	"strings"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
)

/*
	Configure_Middlewares

Sets up the middleware stack for the given router.

Middlewares included:
- Custom Logging Middleware: Logs requests using a custom log formatter.
- CORS Middleware: Allows CORS for all origins (replace this with specific origins).
- RealIP Middleware: Gets the client's real public IP address from the request headers.
- StripSlashes Middleware: Strips slashes to no slash URL versions.
- Panic Recovery Middleware: Recovers from panics and returns a 500 error.

Parameters:

- r: The chi.Mux router to configure middlewares for.
*/
func Configure(r *chi.Mux) {

	// CORS Middleware: the standard deployment (Next.js server proxying /api/* to this backend,
	// see frontend/next.config.ts rewrites) is same-origin from the browser's perspective and
	// never hits this at all - CORS only applies to genuine cross-origin requests. Wildcard +
	// AllowCredentials is also rejected by browsers per spec, so this must be an explicit list;
	// Auth.AllowedOrigins lets an admin opt a specific extra origin in if they're calling the API
	// directly from somewhere else. Router is rebuilt on every config change (see NewRouter), so
	// this always reflects the current config.
	allowedOrigins := config.Current.Auth.AllowedOrigins
	cors := cors.New(cors.Options{
		AllowOriginFunc: func(origin string) bool {
			return slices.Contains(allowedOrigins, origin)
		},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Api-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	r.Use(cors.Handler)

	// RealIP Middleware: Get the client's real public IP address from the request headers
	r.Use(middleware.RealIP)

	// StripSlashes Middleware: Strip slashes to no slash URL versions
	r.Use(middleware.StripSlashes)

	// Custom middleware to remove extra slashes
	r.Use(removeExtraSlashes)

	// Logging Middleware: Custom logging middleware
	r.Use(LoggingMiddleware)

	// Middleware for recovering from panics
	r.Use(middleware.Recoverer)
}

func removeExtraSlashes(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.ReplaceAll(r.URL.Path, "//", "/")
		next.ServeHTTP(w, r)
	})
}
