package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/MDYUSU/mini-jira/internal/repository"
	"github.com/MDYUSU/mini-jira/internal/service"
)

type TaskHandler struct {
	tasks *service.TaskService
}

func NewTaskHandler(tasks *service.TaskService) *TaskHandler {
	return &TaskHandler{tasks: tasks}
}

func (h *TaskHandler) ListByProject(w http.ResponseWriter, r *http.Request) {
	projectID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	q := r.URL.Query()

	page, err := optionalInt(q.Get("page"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "page must be a number")
		return
	}
	limit, err := optionalInt(q.Get("limit"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "limit must be a number")
		return
	}

	result, err := h.tasks.ListByProjectPage(r.Context(), projectID, service.TaskListQuery{
		Status:   q.Get("status"),
		Priority: q.Get("priority"),
		Assignee: q.Get("assignee"),
		Label:    q.Get("label"),
		Search:   q.Get("search"),
		TopLevel: q.Get("top_level") == "true",
		Sort:     q.Get("sort"),
		Order:    q.Get("order"),
		Page:     page,
		Limit:    limit,
	})
	if err != nil {
		writeTaskError(w, err, "list tasks", "project not found")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	task, err := h.tasks.Get(r.Context(), id)
	if err != nil {
		writeTaskError(w, err, "get task", "task not found")
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) ListActivity(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	logs, err := h.tasks.ListActivity(r.Context(), id)
	if err != nil {
		writeTaskError(w, err, "list activity", "task not found")
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	var in service.CreateTaskInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	task, err := h.tasks.Create(r.Context(), projectID, in)
	if err != nil {
		writeTaskError(w, err, "create task", "project not found")
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	var in service.UpdateTaskInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	task, err := h.tasks.Update(r.Context(), id, in)
	if err != nil {
		writeTaskError(w, err, "update task", "task not found")
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	if err := h.tasks.Delete(r.Context(), id); err != nil {
		writeTaskError(w, err, "delete task", "task not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func optionalInt(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}
	return strconv.Atoi(raw)
}

func writeTaskError(w http.ResponseWriter, err error, op, notFound string) {
	var ve service.ValidationError
	switch {
	case errors.As(err, &ve):
		writeError(w, http.StatusBadRequest, ve.Message)
	case errors.Is(err, service.ErrViewerReadOnly):
		writeError(w, http.StatusForbidden, service.ErrViewerReadOnly.Error())
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, notFound)
	case errors.Is(err, repository.ErrCheck), errors.Is(err, repository.ErrForeignKey):
		writeError(w, http.StatusBadRequest, "invalid value")
	default:
		log.Printf("%s: %v", op, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}