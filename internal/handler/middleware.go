package handler

import (
	"net/http"
	"strings"

	"github.com/MDYUSU/mini-jira/internal/auth"
)

func RequireAuth(tokens *auth.TokenManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		raw, ok := strings.CutPrefix(header, "Bearer ")
		raw = strings.TrimSpace(raw)
		if !ok || raw == "" {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}

		userID, err := tokens.Parse(raw)
		if err != nil {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		next.ServeHTTP(w, r.WithContext(auth.WithUserID(r.Context(), userID)))
	})
}
