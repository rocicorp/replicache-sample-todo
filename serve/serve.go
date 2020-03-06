package serve

import (
	"net/http"
)

// Handler implements the Zeit Now entrypoint for our server.
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello"))
}
