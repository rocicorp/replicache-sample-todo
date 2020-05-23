package list

import (
	"encoding/json"
	"fmt"
	"io"

	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/model/list"
	"roci.dev/replicache-sample-todo/serve/util/errs"
)

type ListCreateInput struct {
	ID int `json:"id"`
}

func Create(r io.Reader, exec db.ExecFunc, userID int) error {
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
	if has {
		return errs.NewBadRequestError(fmt.Sprintf("specified list already exists: %d", input.ID))
	}

	err = list.Create(exec, list.List{
		ID:          input.ID,
		OwnerUserID: userID,
	})
	if err != nil {
		return err
	}
	return nil
}
