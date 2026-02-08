package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"
)

func Logging(customMessage string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ts := time.Now().Format(time.RFC3339)
			method := strings.ToUpper(r.Method)
			path := r.URL.Path

			rid := GetRequestID(r)
			if rid != "" {
				log.Printf("%s %s %s %s rid=%s", ts, method, path, customMessage, rid)
			} else {
				log.Printf("%s %s %s %s", ts, method, path, customMessage)
			}

			next.ServeHTTP(w, r)
		})
	}
}
