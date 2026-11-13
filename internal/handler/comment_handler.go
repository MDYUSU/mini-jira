package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/MDYUSU/mini-jira/internal/repository"
	"github.com/MDYUSU/mini-jira/internal/service"
)

type CommentHandler struct {
	comments *service.CommentService
}

func NewCommentHandler(comments *service.CommentService) *CommentHandler {
	return &CommentHandler{comments: comments}
}

func (h *CommentHandler) ListByTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	comments, err := h.comments.ListByTask(r.Context(), taskID)
	if err != nil {
		writeCommentError(w, err, "list comments", "task not found")
		return
	}
	writeJSON(w, http.StatusOK, comments)
}

func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	var in service.CreateCommentInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	comment, err := h.comments.Create(r.Context(), taskID, in)
	if err != nil {
		writeCommentError(w, err, "create comment", "task not found")
		return
	}
	writeJSON(w, http.StatusCreated, comment)
}

func (h *CommentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid comment id")
		return
	}
	var in service.UpdateCommentInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	comment, err := h.comments.Update(r.Context(), id, in)
	if err != nil {
		writeCommentError(w, err, "update comment", "comment not found")
		return
	}
	writeJSON(w, http.StatusOK, comment)
}

func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid comment id")
		return
	}
	actorID, err := parseQueryID(r, "actor_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "actor_id query parameter is required")
		return
	}
	if err := h.comments.Delete(r.Context(), id, actorID); err != nil {
		writeCommentError(w, err, "delete comment", "comment not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeCommentError(w http.ResponseWriter, err error, op, notFound string) {
	var ve service.ValidationError
	switch {
	case errors.As(err, &ve):
		writeError(w, http.StatusBadRequest, ve.Message)
	case errors.Is(err, service.ErrNotAuthor):
		writeError(w, http.StatusForbidden, service.ErrNotAuthor.Error())
	case errors.Is(err, service.ErrCannotDeleteComment):
		writeError(w, http.StatusForbidden, service.ErrCannotDeleteComment.Error())
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, notFound)
	case errors.Is(err, repository.ErrCheck):
		writeError(w, http.StatusBadRequest, "invalid value")
	default:
		log.Printf("%s: %v", op, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}