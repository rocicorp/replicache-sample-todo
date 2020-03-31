package user

import (
	"encoding/json"
	"net/http"

	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/model/user"
	"roci.dev/replicache-sample-todo/serve/types"
	"roci.dev/replicache-sample-todo/serve/util/httperr"
)

func Login(w http.ResponseWriter, r *http.Request, db *db.DB) bool {
	var input types.LoginInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		httperr.ClientError(w, err.Error())
		return false
	}

	if input.Email == "" {
		httperr.ClientError(w, "email field is required")
		return false
	}

	id, err := user.FindByEmail(db, input.Email)
	if err != nil {
		httperr.ServerError(w, err.Error())
		return false
	}

	if id == 0 {
		id, err = user.Create(db, input.Email)
		if err != nil {
			httperr.ServerError(w, err.Error())
			return false
		}
	}

	err = json.NewEncoder(w).Encode(types.LoginOutput{
		Id: id,
	})
	if err != nil {
		httperr.ServerError(w, err.Error())
		return false
	}

	return true
}