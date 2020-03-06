package serve

import (
	"fmt"
	"net/http"

	"roci.dev/replicache-sample-todo/serve/schema"
)

// Handler implements the Zeit Now entrypoint for our server.
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, calling schema()\n\n"))
	err := schema.Create()
	w.Write([]byte(fmt.Sprintf("err: %s", err)))
}
