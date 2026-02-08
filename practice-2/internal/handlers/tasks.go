package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"practice-2/internal/store"
)

type TasksHandler struct {
	Store *store.Store
}

type errResp struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func parseID(r *http.Request) (id int, present bool, invalid bool) {
	raw := r.URL.Query().Get("id")
	if raw == "" {
		return 0, false, false
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 0, true, true
	}
	return n, true, false
}

func parseDoneFilter(r *http.Request) (*bool, error) {
	raw := r.URL.Query().Get("done")
	if raw == "" {
		return nil, nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (h *TasksHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, present, invalid := parseID(r)
	if present {
		if invalid {
			writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid id"})
			return
		}
		t, err := h.Store.Get(id)
		if err == store.ErrNotFound {
			writeJSON(w, http.StatusNotFound, errResp{Error: "task not found"})
			return
		}
		writeJSON(w, http.StatusOK, t)
		return
	}

	doneFilter, err := parseDoneFilter(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid done filter"})
		return
	}

	tasks := h.Store.List(doneFilter)
	writeJSON(w, http.StatusOK, tasks)
}

func (h *TasksHandler) Post(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid json"})
		return
	}

	t, err := h.Store.Create(req.Title)
	switch err {
	case nil:
		writeJSON(w, http.StatusCreated, t)
	case store.ErrBadTitle:
		writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid title"})
	case store.ErrTitleTooLong:
		writeJSON(w, http.StatusBadRequest, errResp{Error: "title too long"})
	default:
		writeJSON(w, http.StatusInternalServerError, errResp{Error: "internal error"})
	}
}

func (h *TasksHandler) Patch(w http.ResponseWriter, r *http.Request) {
	id, present, invalid := parseID(r)
	if !present || invalid {
		writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid id"})
		return
	}

	var req struct {
		Done *bool `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid json"})
		return
	}
	if req.Done == nil {
		writeJSON(w, http.StatusBadRequest, errResp{Error: "done must be boolean"})
		return
	}

	if err := h.Store.UpdateDone(id, *req.Done); err == store.ErrNotFound {
		writeJSON(w, http.StatusNotFound, errResp{Error: "task not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"updated": true})
}

func (h *TasksHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, present, invalid := parseID(r)
	if !present || invalid {
		writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid id"})
		return
	}
	if err := h.Store.Delete(id); err == store.ErrNotFound {
		writeJSON(w, http.StatusNotFound, errResp{Error: "task not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}
