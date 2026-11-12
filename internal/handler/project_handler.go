package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/MDYUSU/mini-jira/internal/repository"
	"github.com/MDYUSU/mini-jira/internal/service"
)

type ProjectHandler struct {
	projects *service.ProjectService
}

func NewProjectHandler(projects *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projects: projects}
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	project, err := h.projects.Get(r.Context(), id)
	if err != nil {
		writeProjectError(w, err, "get project", "project not found", "")
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization id")
		return
	}
	var in service.CreateProjectInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	project, err := h.projects.Create(r.Context(), orgID, in)
	if err != nil {
		writeProjectError(w, err, "create project", "organization not found", "project key already exists in this organization")
		return
	}
	writeJSON(w, http.StatusCreated, project)
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	var in service.UpdateProjectInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	project, err := h.projects.Update(r.Context(), id, in)
	if err != nil {
		writeProjectError(w, err, "update project", "project not found", "project key already exists in this organization")
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	if err := h.projects.Delete(r.Context(), id); err != nil {
		writeProjectError(w, err, "delete project", "project not found", "")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProjectHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	projectID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	var in service.AddMemberInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	member, err := h.projects.AddMember(r.Context(), projectID, in)
	if err != nil {
		writeProjectError(w, err, "add member", "project not found", "user is already a member of this project")
		return
	}
	writeJSON(w, http.StatusCreated, member)
}

func (h *ProjectHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	projectID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	userID, err := parseID(r, "userId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var in service.UpdateMemberInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	member, err := h.projects.UpdateMemberRole(r.Context(), projectID, userID, in)
	if err != nil {
		writeProjectError(w, err, "update member role", "member not found", "")
		return
	}
	writeJSON(w, http.StatusOK, member)
}

func (h *ProjectHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	projectID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	userID, err := parseID(r, "userId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	if err := h.projects.RemoveMember(r.Context(), projectID, userID); err != nil {
		writeProjectError(w, err, "remove member", "member not found", "")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeProjectError(w http.ResponseWriter, err error, op, notFound, conflict string) {
	var ve service.ValidationError
	switch {
	case errors.As(err, &ve):
		writeError(w, http.StatusBadRequest, ve.Message)
	case errors.Is(err, service.ErrLastOwner):
		writeError(w, http.StatusConflict, service.ErrLastOwner.Error())
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, notFound)
	case errors.Is(err, repository.ErrConflict) && conflict != "":
		writeError(w, http.StatusConflict, conflict)
	case errors.Is(err, repository.ErrCheck):
		writeError(w, http.StatusBadRequest, "invalid value")
	default:
		log.Printf("%s: %v", op, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}