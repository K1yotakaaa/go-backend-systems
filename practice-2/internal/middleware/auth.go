package middleware

import (
	"encoding/json"
	"net/http"
)

type errResp struct {
	Error string `json:"error"`
}

func APIKey(validKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-API-KEY") != validKey {
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(errResp{Error: "unauthorized"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
