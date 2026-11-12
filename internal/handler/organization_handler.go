package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/MDYUSU/mini-jira/internal/repository"
	"github.com/MDYUSU/mini-jira/internal/service"
)

type OrganizationHandler struct {
	orgs *service.OrganizationService
}

func NewOrganizationHandler(orgs *service.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{orgs: orgs}
}

func (h *OrganizationHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization id")
		return
	}
	org, err := h.orgs.Get(r.Context(), id)
	if err != nil {
		writeOrgError(w, err, "get org")
		return
	}
	writeJSON(w, http.StatusOK, org)
}

func (h *OrganizationHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization id")
		return
	}
	projects, err := h.orgs.ListProjects(r.Context(), id)
	if err != nil {
		writeOrgError(w, err, "list org projects")
		return
	}
	writeJSON(w, http.StatusOK, projects)
}

func (h *OrganizationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in service.CreateOrganizationInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	org, err := h.orgs.Create(r.Context(), in)
	if err != nil {
		writeOrgError(w, err, "create org")
		return
	}
	writeJSON(w, http.StatusCreated, org)
}

func writeOrgError(w http.ResponseWriter, err error, op string) {
	var ve service.ValidationError
	switch {
	case errors.As(err, &ve):
		writeError(w, http.StatusBadRequest, ve.Message)
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "organization not found")
	default:
		log.Printf("%s: %v", op, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}