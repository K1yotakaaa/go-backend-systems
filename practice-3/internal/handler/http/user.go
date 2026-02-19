package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"practice-3/internal/usecase"
	"practice-3/pkg/modules"
)

var _ = modules.User{}

type UserHandler struct {
	uc *usecase.UserUsecase
}

func NewUserHandler(uc *usecase.UserUsecase) *UserHandler { return &UserHandler{uc: uc} }

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// Health godoc
// @Summary Healthcheck
// @Description Returns API status
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *UserHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

// GetUsers godoc
// @Summary Get all users
// @Description Get list of users (supports pagination)
// @Tags users
// @Security ApiKeyAuth
// @Produce json
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} modules.User
// @Failure 500 {object} map[string]string
// @Router /users [get]
func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	users, err := h.uc.GetUsers(limit, offset)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, users)
}

// GetUserByID godoc
// @Summary Get user by id
// @Tags users
// @Security ApiKeyAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} modules.User
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id} [get]
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}

	u, err := h.uc.GetUserByID(id)
	if err != nil {
		if usecase.IsNotFound(err) {
			writeJSON(w, 404, map[string]string{"error": "user not found"})
			return
		}
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, 200, u)
}

type CreateUserRequest struct {
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Age      int     `json:"age"`
	Password *string `json:"password,omitempty"`
}

// CreateUser godoc
// @Summary Create user
// @Tags users
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param body body CreateUserRequest true "Create user payload"
// @Success 201 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Router /users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	if req.Name == "" || req.Email == "" {
		writeJSON(w, 400, map[string]string{"error": "name and email are required"})
		return
	}

	id, err := h.uc.CreateUser(req.Name, req.Email, req.Age, req.Password)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}

type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// UpdateUser godoc
// @Summary Update user
// @Tags users
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param body body UpdateUserRequest true "Update user payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}

	if err := h.uc.UpdateUser(id, req.Name, req.Email, req.Age); err != nil {
		if usecase.IsNotFound(err) {
			writeJSON(w, 404, map[string]string{"error": "user not found"})
			return
		}
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, 200, map[string]string{"status": "updated"})
}

// DeleteUser godoc
// @Summary Delete user
// @Tags users
// @Security ApiKeyAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}

	ra, err := h.uc.DeleteUser(id)
	if err != nil {
		if usecase.IsNotFound(err) {
			writeJSON(w, 404, map[string]string{"error": "user not found"})
			return
		}
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, 200, map[string]any{"status": "deleted", "rows_affected": ra})
}
