package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/MDYUSU/mini-jira/internal/repository"
)

type ProjectHandler struct {
	projects *repository.ProjectRepository
}

func NewProjectHandler(projects *repository.ProjectRepository) *ProjectHandler {
	return &ProjectHandler{projects: projects}
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	project, err := h.projects.GetDetail(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		log.Printf("get project %d: %v", id, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, project)
}