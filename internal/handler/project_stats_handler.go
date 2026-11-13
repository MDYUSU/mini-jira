package handler

import "net/http"

func (h *ProjectHandler) Stats(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	stats, err := h.projects.Stats(r.Context(), id)
	if err != nil {
		writeProjectError(w, err, "project stats", "project not found", "")
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (h *ProjectHandler) Workload(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	list, err := h.projects.Workload(r.Context(), id)
	if err != nil {
		writeProjectError(w, err, "project workload", "project not found", "")
		return
	}
	writeJSON(w, http.StatusOK, list)
}
