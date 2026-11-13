package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/MDYUSU/mini-jira/internal/repository"
	"github.com/MDYUSU/mini-jira/internal/service"
)

type LabelHandler struct {
	labels *service.LabelService
}

func NewLabelHandler(labels *service.LabelService) *LabelHandler {
	return &LabelHandler{labels: labels}
}

func (h *LabelHandler) ListByProject(w http.ResponseWriter, r *http.Request) {
	projectID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	labels, err := h.labels.ListByProject(r.Context(), projectID)
	if err != nil {
		writeLabelError(w, err, "list labels", "project not found", "")
		return
	}
	writeJSON(w, http.StatusOK, labels)
}

func (h *LabelHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	var in service.CreateLabelInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	label, err := h.labels.Create(r.Context(), projectID, in)
	if err != nil {
		writeLabelError(w, err, "create label", "project not found", "a label with this name already exists in this project")
		return
	}
	writeJSON(w, http.StatusCreated, label)
}

func (h *LabelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid label id")
		return
	}
	if err := h.labels.Delete(r.Context(), id); err != nil {
		writeLabelError(w, err, "delete label", "label not found", "")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *LabelHandler) Attach(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	var in service.AttachLabelInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	labels, err := h.labels.Attach(r.Context(), taskID, in)
	if err != nil {
		writeLabelError(w, err, "attach label", "task not found", "label is already attached to this task")
		return
	}
	writeJSON(w, http.StatusCreated, labels)
}

func (h *LabelHandler) Detach(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	labelID, err := parseID(r, "labelId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid label id")
		return
	}
	actorID, err := parseQueryID(r, "actor_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "actor_id query parameter is required")
		return
	}
	if err := h.labels.Detach(r.Context(), taskID, labelID, actorID); err != nil {
		writeLabelError(w, err, "detach label", "task not found", "")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeLabelError(w http.ResponseWriter, err error, op, notFound, conflict string) {
	var ve service.ValidationError
	switch {
	case errors.As(err, &ve):
		writeError(w, http.StatusBadRequest, ve.Message)
	case errors.Is(err, service.ErrViewerReadOnly):
		writeError(w, http.StatusForbidden, service.ErrViewerReadOnly.Error())
	case errors.Is(err, service.ErrLabelNotAttached):
		writeError(w, http.StatusNotFound, service.ErrLabelNotAttached.Error())
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, notFound)
	case errors.Is(err, repository.ErrConflict) && conflict != "":
		writeError(w, http.StatusConflict, conflict)
	default:
		log.Printf("%s: %v", op, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}