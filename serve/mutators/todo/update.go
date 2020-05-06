package todo

import (
	"encoding/json"
	"io"

	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/model/todo"
	"roci.dev/replicache-sample-todo/serve/util/errs"
)

type UpdateInput struct {
	ID       int      `json:"id"`
	Text     *string  `json:"text,omitempty"`
	Complete *bool    `json:"complete,omitempty"`
	Order    *float64 `json:"order,omitempty"`
}

func Update(r io.Reader, db *db.DB, userID int) error {
	var input UpdateInput
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
		return errs.NewBadRequestError("specified todo not found")
	}

	if got.OwnerUserID != userID {
		return errs.NewUnauthorizedError("access unauthorized")
	}

	err = todo.Update(db, input.ID, input.Complete, input.Order, input.Text)
	if err != nil {
		return err
	}
	return nil
}
