package list

import (
	"encoding/json"
	"io"

	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/model/list"
	"roci.dev/replicache-sample-todo/serve/util/errs"
)

type ListDeleteInput struct {
	ID int `json:"id"`
}

func Delete(r io.Reader, exec db.ExecFunc, userID int) error {
	var input ListCreateInput
	err := json.NewDecoder(r).Decode(&input)
	if err != nil {
		return errs.NewBadRequestError(err.Error())
	}
	if input.ID == 0 {
		return errs.NewBadRequestError("id field is required")
	}

	_, has, err := list.Get(exec, input.ID)
	if err != nil {
		return err
	}
	if !has {
		return errs.NewBadRequestError("list not found")
	}

	err = list.Delete(exec, input.ID)
	if err != nil {
		return err
	}
	return nil
}
