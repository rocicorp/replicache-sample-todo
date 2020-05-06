package serve

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/handlers/clientview"
	"roci.dev/replicache-sample-todo/serve/handlers/mutator"
	userhandler "roci.dev/replicache-sample-todo/serve/handlers/user"
	"roci.dev/replicache-sample-todo/serve/model/schema"
	"roci.dev/replicache-sample-todo/serve/model/user"
	"roci.dev/replicache-sample-todo/serve/mutators/list"
	"roci.dev/replicache-sample-todo/serve/mutators/todo"
	"roci.dev/replicache-sample-todo/serve/util/httperr"
)

const (
	awsAccessKeyId     = "REPLICANT_AWS_ACCESS_KEY_ID"
	awsSecretAccessKey = "REPLICANT_AWS_SECRET_ACCESS_KEY"
	awsRegion          = "us-west-2"
	schemaVersion      = 2
)

// Handler implements the Zeit Now entrypoint for our server.
func Handler(w http.ResponseWriter, r *http.Request) {
	name, err := dbName()
	if err != nil {
		httperr.ServerError(w, err.Error())
		return
	}
	db := db.New()

	err = schema.Create(db, name)
	if err != nil {
		httperr.ServerError(w, err.Error())
		return
	}

	db.Use(name)

	switch r.URL.Path {
	case "/serve/login":
		userhandler.Login(w, r, db)
		return
	}

	userID := authenticate(db, w, r)
	if userID == 0 {
		return
	}

	switch r.URL.Path {
	case "/serve/replicache-batch":
		clientview.Handle(w, r, db, userID)
	case "/serve/replicache-client-view":
		clientview.Handle(w, r, db, userID)
	case "/serve/list-create":
		mutator.Handle(w, func() error {
			return list.Create(r.Body, db, userID)
		})
	case "/serve/todo-create":
		mutator.Handle(w, func() error {
			return todo.Create(r.Body, db, userID)
		})
	case "/serve/todo-update":
		mutator.Handle(w, func() error {
			return todo.Update(r.Body, db, userID)
		})
	case "/serve/todo-delete":
		mutator.Handle(w, func() error {
			return todo.Delete(r.Body, db, userID)
		})
	default:
		httperr.ClientError(w, fmt.Sprintf("Unknown path: %s", r.URL.Path))
		return
	}

	if err != nil {
		httperr.ServerError(w, err.Error())
	}
}

func authenticate(db *db.DB, w http.ResponseWriter, r *http.Request) (userID int) {
	s := r.Header.Get("Authorization")
	if s == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Authentication required"))
		return 0
	}
	userID, err := strconv.Atoi(s)
	if err != nil || userID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Authorization header"))
		return 0
	}

	ok, err := user.Has(db, userID)
	if err != nil || !ok {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Authentication failed"))
		return 0
	}

	return userID
}

func dbName() (string, error) {
	n := "REPLICANT_SAMPLE_TODO_ENV"
	env := os.Getenv(n)
	if env == "" {
		return "", fmt.Errorf("Required environment variable %s not found", n)
	} else {
		return fmt.Sprintf("replicache_sample_todo__%s", env), nil
	}
}
