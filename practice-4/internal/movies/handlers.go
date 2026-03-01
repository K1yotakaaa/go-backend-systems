package movies

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	Repo Repo
}

type upsertReq struct {
	Title  string `json:"title"`
	Genre  string `json:"genre"`
	Budget int64  `json:"budget"`
}

func (h Handlers) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.get)
		r.Put("/", h.update)
		r.Delete("/", h.delete)
	})

	return r
}

func (h Handlers) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.Repo.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, 200, items)
}

func (h Handlers) get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "bad id", 400)
		return
	}

	m, ok, err := h.Repo.Get(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if !ok {
		http.Error(w, "not found", 404)
		return
	}

	writeJSON(w, 200, m)
}

func (h Handlers) create(w http.ResponseWriter, r *http.Request) {
	var req upsertReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	if req.Title == "" || req.Genre == "" {
		http.Error(w, "title and genre required", 400)
		return
	}

	m, err := h.Repo.Create(r.Context(), req.Title, req.Genre, req.Budget)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, 201, m)
}

func (h Handlers) update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "bad id", 400)
		return
	}

	var req upsertReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	if req.Title == "" || req.Genre == "" {
		http.Error(w, "title and genre required", 400)
		return
	}

	m, ok, err := h.Repo.Update(r.Context(), id, req.Title, req.Genre, req.Budget)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if !ok {
		http.Error(w, "not found", 404)
		return
	}

	writeJSON(w, 200, m)
}

func (h Handlers) delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "bad id", 400)
		return
	}

	ok, err := h.Repo.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if !ok {
		http.Error(w, "not found", 404)
		return
	}

	w.WriteHeader(204)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}