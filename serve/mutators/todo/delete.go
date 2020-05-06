package todo

import (
	"encoding/json"
	"io"

	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/model/todo"
	"roci.dev/replicache-sample-todo/serve/util/errs"
)

type Input struct {
	ID int `json:"id"`
}

func Delete(r io.Reader, db *db.DB, userID int) error {
	var input Input
	err := json.NewDecoder(r).Decode(&input)
	if err != nil {
		return errs.NewBadRequestError(err.Error())
	}
	if input.ID == 0 {
		return errs.NewBadRequestError("id field is required")
	}

	got, has, err := todo.Get(db, input.ID)
	if err != nil {
		return err
	}
	if !has {
		return errs.NewBadRequestError("todo not found")
	}

	if got.OwnerUserID != userID {
		return errs.NewUnauthorizedError("access unauthorized")
	}

	err = todo.Delete(db, input.ID)
	if err != nil {
		return err
	}
	return nil
}
