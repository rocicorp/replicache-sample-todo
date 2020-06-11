package user

import (
	"encoding/json"
	"net/http"

	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/model/list"
	"roci.dev/replicache-sample-todo/serve/model/user"
	"roci.dev/replicache-sample-todo/serve/util/httperr"
)

type LoginInput struct {
	Email string `json:"email"`
}

type LoginOutput struct {
	Id int `json:"id"`
}

func Login(w http.ResponseWriter, r *http.Request, d *db.DB) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var input LoginInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		httperr.ClientError(w, err.Error())
		return
	}

	if input.Email == "" {
		httperr.ClientError(w, "email field is required")
		return
	}

	var id int
	_, err = d.Transact(func(exec db.ExecFunc) bool {
		var err error
		id, err = user.FindByEmail(exec, input.Email)
		if err != nil {
			httperr.ServerError(w, err.Error())
			return false
		}

		if id == 0 {
			id, err = user.Create(exec, input.Email)
			if err != nil {
				httperr.ServerError(w, err.Error())
				return false
			}
			listID, err := list.GetMax(exec)
			if err != nil {
				httperr.ServerError(w, err.Error())
				return false
			}
			listID++
			err = list.Create(exec, list.List{
				ID:          listID,
				OwnerUserID: id,
			})
			if err != nil {
				httperr.ServerError(w, err.Error())
				return false
			}
		}
		return true
	})

	if err != nil {
		httperr.ServerError(w, err.Error())
		return
	}

	err = json.NewEncoder(w).Encode(LoginOutput{
		Id: id,
	})
	if err != nil {
		httperr.ServerError(w, err.Error())
	}
}
