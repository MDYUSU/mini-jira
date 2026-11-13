package handler

import "net/http"

func (h *TaskHandler) Tree(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	tree, err := h.tasks.GetTree(r.Context(), id)
	if err != nil {
		writeTaskError(w, err, "get task tree", "task not found")
		return
	}
	writeJSON(w, http.StatusOK, tree)
}
