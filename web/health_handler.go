package web

import (
	"fmt"
	"net/http"
)

// HealthHandler is standart health check for appcop
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprint(w, "OK")
	if err != nil {
		return
	}
}
