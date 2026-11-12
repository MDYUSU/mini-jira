package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/MDYUSU/mini-jira/internal/repository"
)

type OrganizationHandler struct {
	orgs     *repository.OrganizationRepository
	projects *repository.ProjectRepository
}

func NewOrganizationHandler(orgs *repository.OrganizationRepository, projects *repository.ProjectRepository) *OrganizationHandler {
	return &OrganizationHandler{orgs: orgs, projects: projects}
}

func (h *OrganizationHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization id")
		return
	}

	org, err := h.orgs.GetByID(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "organization not found")
		return
	}
	if err != nil {
		log.Printf("get org %d: %v", id, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
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

	if _, err := h.orgs.GetByID(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "organization not found")
			return
		}
		log.Printf("get org %d: %v", id, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	projects, err := h.projects.ListByOrg(r.Context(), id)
	if err != nil {
		log.Printf("list projects for org %d: %v", id, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, projects)
}