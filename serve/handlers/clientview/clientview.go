package clientview

import (
	"encoding/json"
	"fmt"
	"net/http"

	servetypes "roci.dev/diff-server/serve/types"
	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/model/list"
	"roci.dev/replicache-sample-todo/serve/model/todo"
	"roci.dev/replicache-sample-todo/serve/util/httperr"
)

func Handle(w http.ResponseWriter, r *http.Request, db *db.DB, userID int) {
	var lists []list.List
	var todos []todo.Todo
	_, err := db.Transact(func() bool {
		var err error
		lists, err = list.GetByUser(db, userID)
		if err != nil {
			httperr.ServerError(w, err.Error())
			return false
		}
		todos, err = todo.GetByUser(db, userID)
		if err != nil {
			httperr.ServerError(w, err.Error())
			return false
		}
		return true
	})
	if err != nil {
		httperr.ServerError(w, err.Error())
		return
	}
	out := servetypes.ClientViewResponse{
		ClientView:     map[string]interface{}{},
		LastMutationID: 0,
	}
	for _, l := range lists {
		out.ClientView[fmt.Sprintf("/list/%d", l.ID)] = list.List(l)
	}
	for _, t := range todos {
		out.ClientView[fmt.Sprintf("/todo/%d", t.ID)] = todo.Todo(t)
	}
	err = json.NewEncoder(w).Encode(out)
	if err != nil {
		httperr.ServerError(w, err.Error())
		return
	}
}
