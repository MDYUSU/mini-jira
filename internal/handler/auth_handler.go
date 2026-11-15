package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/MDYUSU/mini-jira/internal/repository"
	"github.com/MDYUSU/mini-jira/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var in service.RegisterInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	res, err := h.auth.Register(r.Context(), in)
	if err != nil {
		writeAuthError(w, err, "register")
		return
	}
	writeJSON(w, http.StatusCreated, res)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var in service.LoginInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	res, err := h.auth.Login(r.Context(), in)
	if err != nil {
		writeAuthError(w, err, "login")
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func writeAuthError(w http.ResponseWriter, err error, op string) {
	var ve service.ValidationError
	switch {
	case errors.As(err, &ve):
		writeError(w, http.StatusBadRequest, ve.Message)
	case errors.Is(err, service.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, service.ErrInvalidCredentials.Error())
	case errors.Is(err, repository.ErrConflict):
		writeError(w, http.StatusConflict, "email already in use")
	default:
		log.Printf("%s: %v", op, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
