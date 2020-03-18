package clientview

import (
	"encoding/json"
	"fmt"
	"net/http"

	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/model/list"
	"roci.dev/replicache-sample-todo/serve/model/todo"
	"roci.dev/replicache-sample-todo/serve/types"
	"roci.dev/replicache-sample-todo/serve/util/httperr"
)

func Handle(w http.ResponseWriter, r *http.Request, db *db.DB, userID int) {
	_, err := db.Transact(func() (commit bool) {
		lists, err := list.GetByUser(db, userID)
		if err != nil {
			httperr.ServerError(w, err.Error())
			return false
		}
		todos, err := todo.GetByUser(db, userID)
		if err != nil {
			httperr.ServerError(w, err.Error())
			return false
		}
		out := types.ClientViewOutput{
			View: map[string]interface{}{},
		}
		for _, l := range lists {
			out.View[fmt.Sprintf("/list/%d", l.ID)] = types.TodoList(l)
		}
		for _, t := range todos {
			out.View[fmt.Sprintf("/todo/%d", t.ID)] = types.Todo(t)
		}
		err = json.NewEncoder(w).Encode(out)
		if err != nil {
			httperr.ServerError(w, err.Error())
			return false
		}
		return true
	})

	if err != nil {
		httperr.ServerError(w, err.Error())
	}
}
