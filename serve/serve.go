package serve

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/handlers/clientview"
	"roci.dev/replicache-sample-todo/serve/handlers/todo"
	userhandler "roci.dev/replicache-sample-todo/serve/handlers/user"
	"roci.dev/replicache-sample-todo/serve/model/schema"
	"roci.dev/replicache-sample-todo/serve/model/user"
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

	_, err = db.Transact(func() bool {
		switch r.URL.Path {
		case "/serve/login":
			return userhandler.Login(w, r, db)
		}

		userID := authenticate(db, w, r)
		if userID == 0 {
			return false
		}

		switch r.URL.Path {
		case "/serve/todo-create":
			return todo.Handle(w, r, db, userID)
		case "/serve/client-view":
			return clientview.Handle(w, r, db, userID)
		default:
			httperr.ClientError(w, fmt.Sprintf("Unknown path: %s", r.URL.Path))
			return false
		}
	})

	if err != nil {
		httperr.ServerError(w, err.Error())
		return
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
