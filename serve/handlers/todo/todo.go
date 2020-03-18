package todo

import (
	"encoding/json"
	"fmt"
	"net/http"

	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/model/list"
	"roci.dev/replicache-sample-todo/serve/model/todo"
	"roci.dev/replicache-sample-todo/serve/model/user"
	"roci.dev/replicache-sample-todo/serve/types"
	"roci.dev/replicache-sample-todo/serve/util/httperr"
)

func Handle(w http.ResponseWriter, r *http.Request, db *db.DB, userID int) {
	var input types.TodoCreateInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		httperr.ClientError(w, err.Error())
		return
	}

	if input.ID == 0 {
		httperr.ClientError(w, "id field is required")
		return
	}

	if input.ListID == 0 {
		httperr.ClientError(w, "listID field is required")
		return
	}

	_, err = db.Transact(func() (commit bool) {
		if !ensureUser(w, r, db, userID) {
			return false
		}
		if !ensureList(w, r, db, input.ListID, userID) {
			return false
		}
		if !ensureTodo(w, r, db, input) {
			return false
		}
		return true
	})

	if err != nil {
		httperr.ServerError(w, err.Error())
		return
	}

	return
}

func ensureUser(w http.ResponseWriter, r *http.Request, db *db.DB, userID int) bool {
	has, err := user.Has(db, userID)
	if err != nil {
		httperr.ServerError(w, err.Error())
		return false
	}

	if !has {
		err := user.Create(db, userID)
		if err != nil {
			httperr.ServerError(w, err.Error())
			return false
		}
	}

	return true
}

func ensureList(w http.ResponseWriter, r *http.Request, db *db.DB, listID int, ownerUserID int) bool {
	l, has, err := list.Get(db, listID)
	if err != nil {
		httperr.ServerError(w, err.Error())
		return false
	}
	if has {
		if l.OwnerUserID != ownerUserID {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(fmt.Sprintf("Cannot access specified list")))
			return false
		}
		return true
	}
	l = list.List{
		ID:          listID,
		OwnerUserID: ownerUserID,
	}
	err = list.Create(db, l)
	if err != nil {
		httperr.ServerError(w, err.Error())
		return false
	}
	return true
}

func ensureTodo(w http.ResponseWriter, r *http.Request, db *db.DB, input types.TodoCreateInput) bool {
	has, err := todo.Has(db, input.ID)
	if err != nil {
		httperr.ServerError(w, err.Error())
		return false
	}
	if has {
		httperr.ClientError(w, fmt.Sprintf("Specified todo already exists: %d", input.ID))
		return false
	}
	err = todo.Create(db, todo.Todo(input))
	if err != nil {
		httperr.ServerError(w, err.Error())
		return false
	}
	return true
}
