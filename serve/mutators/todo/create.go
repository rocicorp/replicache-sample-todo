package todo

import (
	"encoding/json"
	"fmt"
	"io"

	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/model/list"
	"roci.dev/replicache-sample-todo/serve/model/todo"
	"roci.dev/replicache-sample-todo/serve/util/errs"
)

type CreateInput todo.Todo

func Create(r io.Reader, db *db.DB, userID int) error {
	var input CreateInput
	err := json.NewDecoder(r).Decode(&input)
	if err != nil {
		return errs.NewBadRequestError(err.Error())
	}
	if input.ID == 0 {
		return errs.NewBadRequestError("id field is required")
	}
	if input.ListID == 0 {
		return errs.NewBadRequestError("listID field is required")
	}

	hasTodo, err := todo.Has(db, input.ID)
	if err != nil {
		return err
	}
	if hasTodo {
		return errs.NewBadRequestError(fmt.Sprintf("specified todo already exists: %d", input.ID))
	}

	list, hasList, err := list.Get(db, input.ListID)
	if err != nil {
		return err
	}
	if !hasList {
		return errs.NewBadRequestError(fmt.Sprintf("specified list does not exist: %d", input.ListID))
	}
	if list.OwnerUserID != userID {
		return errs.NewUnauthorizedError(fmt.Sprintf("cannot access specified list: %d", input.ListID))
	}

	err = todo.Create(db, todo.Todo(input))
	if err != nil {
		return err
	}
	return nil
}
