package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/MDYUSU/mini-jira/internal/repository"
	"github.com/MDYUSU/mini-jira/internal/service"
)

type UserHandler struct {
	users *service.UserService
}

func NewUserHandler(users *service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.users.List(r.Context())
	if err != nil {
		writeUserError(w, err, "list users")
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	user, err := h.users.Get(r.Context(), id)
	if err != nil {
		writeUserError(w, err, "get user")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in service.CreateUserInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	user, err := h.users.Create(r.Context(), in)
	if err != nil {
		writeUserError(w, err, "create user")
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var in service.UpdateUserInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	user, err := h.users.Update(r.Context(), id, in)
	if err != nil {
		writeUserError(w, err, "update user")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	if err := h.users.Delete(r.Context(), id); err != nil {
		writeUserError(w, err, "delete user")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) UpsertProfile(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var in service.ProfileInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	profile, err := h.users.UpsertProfile(r.Context(), id, in)
	if err != nil {
		writeUserError(w, err, "upsert profile")
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func writeUserError(w http.ResponseWriter, err error, op string) {
	var ve service.ValidationError
	switch {
	case errors.As(err, &ve):
		writeError(w, http.StatusBadRequest, ve.Message)
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "user not found")
	case errors.Is(err, repository.ErrConflict):
		writeError(w, http.StatusConflict, "email already in use")
	case errors.Is(err, repository.ErrForeignKey):
		writeError(w, http.StatusConflict, "user still owns organizations or has reported tasks")
	default:
		log.Printf("%s: %v", op, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}