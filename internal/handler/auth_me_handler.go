package handler

import (
	"errors"
	"net/http"

	"github.com/MDYUSU/mini-jira/internal/auth"
	"github.com/MDYUSU/mini-jira/internal/repository"
)

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not logged in")
		return
	}
	user, err := h.auth.Me(r.Context(), userID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusUnauthorized, "user no longer exists")
		return
	}
	if err != nil {
		writeAuthError(w, err, "me")
		return
	}
	writeJSON(w, http.StatusOK, user)
}
