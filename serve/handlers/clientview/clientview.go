package clientview

import (
	"encoding/json"
	"fmt"
	"net/http"

	servetypes "roci.dev/diff-server/serve/types"
	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/model/list"
	"roci.dev/replicache-sample-todo/serve/model/replicache"
	"roci.dev/replicache-sample-todo/serve/model/todo"
	"roci.dev/replicache-sample-todo/serve/util/httperr"
)

type clientViewRequest struct {
	ClientID string `json:"clientID"`
}

func Handle(w http.ResponseWriter, r *http.Request, d *db.DB, userID int) {
	var req clientViewRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		httperr.ClientError(w, err.Error())
		return
	}

	if req.ClientID == "" {
		httperr.ClientError(w, "clientID is required")
		return
	}

	var lists []list.List
	var todos []todo.Todo
	var lastMutationID int64
	_, err = d.Transact(func(exec db.ExecFunc) bool {
		var err error
		lastMutationID, err = replicache.GetMutationID(exec, req.ClientID)
		if err != nil {
			httperr.ServerError(w, err.Error())
			return false
		}
		lists, err = list.GetByUser(exec, userID)
		if err != nil {
			httperr.ServerError(w, err.Error())
			return false
		}
		todos, err = todo.GetByUser(exec, userID)
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
		LastMutationID: uint64(lastMutationID),
	}
	for _, l := range lists {
		out.ClientView[fmt.Sprintf("/list/%d", l.ID)] = l
	}
	for _, t := range todos {
		out.ClientView[fmt.Sprintf("/todo/%d", t.ID)] = t
	}
	err = json.NewEncoder(w).Encode(out)
	if err != nil {
		httperr.ServerError(w, err.Error())
		return
	}
}
