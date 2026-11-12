package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/MDYUSU/mini-jira/internal/repository"
)

type TaskHandler struct {
	tasks    *repository.TaskRepository
	projects *repository.ProjectRepository
}

func NewTaskHandler(tasks *repository.TaskRepository, projects *repository.ProjectRepository) *TaskHandler {
	return &TaskHandler{tasks: tasks, projects: projects}
}

func (h *TaskHandler) ListByProject(w http.ResponseWriter, r *http.Request) {
	projectID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	exists, err := h.projects.Exists(r.Context(), projectID)
	if err != nil {
		log.Printf("check project %d: %v", projectID, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !exists {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}

	tasks, err := h.tasks.ListByProject(r.Context(), projectID)
	if err != nil {
		log.Printf("list tasks for project %d: %v", projectID, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	task, err := h.tasks.GetDetail(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		log.Printf("get task %d: %v", id, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
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

	exists, err := h.tasks.Exists(r.Context(), id)
	if err != nil {
		log.Printf("check task %d: %v", id, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !exists {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}

	logs, err := h.tasks.ListActivity(r.Context(), id)
	if err != nil {
		log.Printf("list activity for task %d: %v", id, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, logs)
}