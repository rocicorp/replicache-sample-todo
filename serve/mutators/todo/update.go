package todo

import (
	"encoding/json"
	"fmt"
	"io"

	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/model/todo"
	"roci.dev/replicache-sample-todo/serve/types"
	"roci.dev/replicache-sample-todo/serve/util/errs"
)

func Update(r io.Reader, db *db.DB, userID int) error {
	var input types.TodoUpdateInput
	err := json.NewDecoder(r).Decode(&input)
	if err != nil {
		return errs.NewBadRequestError(err.Error())
	}
	if input.ID == 0 {
		return errs.NewBadRequestError("id field is required")
	}

	_, has, err := todo.Get(db, input.ID, userID)
	if err != nil {
		return err
	}
	if !has {
		return errs.NewBadRequestError(fmt.Sprintf("specified todo not found: %d", input.ID))
	}

	err = todo.Update(db, input.ID, input.Complete, input.Order, input.Text)
	if err != nil {
		return err
	}
	return nil
}
