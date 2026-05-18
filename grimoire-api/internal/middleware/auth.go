package middleware

import (
	"context"
	"net/http"
	"strings"
)

type tokenKey struct{}

// AuthMiddleware extracts the Firebase JWT from the Authorization header
// and places the raw token in the request context.
// CallerIdentityPort.Identify() is called by the resolver — not here.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		ctx := context.WithValue(r.Context(), tokenKey{}, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TokenFromContext retrieves the raw bearer token placed by AuthMiddleware.
func TokenFromContext(ctx context.Context) string {
	token, _ := ctx.Value(tokenKey{}).(string)
	return token
}
