package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

type ExternalHandler struct {
	Client *http.Client
}

func (h *ExternalHandler) Todos(w http.ResponseWriter, r *http.Request) {
	client := h.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	resp, err := client.Get("https://jsonplaceholder.typicode.com/todos")
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errResp{Error: "failed to fetch external api"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		writeJSON(w, http.StatusBadGateway, errResp{Error: "external api returned non-2xx"})
		return
	}

	var data []any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		writeJSON(w, http.StatusBadGateway, errResp{Error: "failed to parse external json"})
		return
	}

	writeJSON(w, http.StatusOK, data)
}
