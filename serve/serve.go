package serve

import (
	"net/http"

	"roci.dev/replicache-sample-todo/serve/schema"
)

// Handler implements the Zeit Now entrypoint for our server.
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, calling schema()"))
	err := schema.create()
	fmt.Println("err", err)
}
