package batch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/model/replicache"
	"roci.dev/replicache-sample-todo/serve/mutators/list"
	"roci.dev/replicache-sample-todo/serve/mutators/todo"
	"roci.dev/replicache-sample-todo/serve/util/errs"
	"roci.dev/replicache-sample-todo/serve/util/httperr"
)

type batchRequest struct {
	ClientID  string     `json:"clientID"`
	Mutations []mutation `json:"mutations"`
}

type mutation struct {
	// NOTE: Replicache protocol specifies this as uint64, but
	// MySQL only has int64.
	ID   int64           `json:"id"`
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}

type batchResponse struct {
	MutationInfos []mutationInfo `json:"mutationInfos"`
}

type mutationInfo struct {
	ID    int64  `json:"id"`
	Error string `json:"error"`
}

// Handle implements the Replicache batch upload endpoint. It processes zero or more
// mutations specified by the request body. The response is purely informational.
func Handle(w http.ResponseWriter, r *http.Request, d *db.DB, userID int) {
	var req batchRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		httperr.ClientError(w, err.Error())
		return
	}

	if req.ClientID == "" {
		httperr.ClientError(w, "clientID is required")
		return
	}

	infos := []mutationInfo{}
	stop := false
	var internalError error
	for _, m := range req.Mutations {
		var err error
		_, txErr := d.Transact(func(exec db.ExecFunc) bool {
			var currentMutationID, expectedMutationID int64

			currentMutationID, internalError = replicache.GetMutationID(exec, req.ClientID)
			if internalError != nil {
				stop = true
				return false
			}

			expectedMutationID = currentMutationID + 1

			if m.ID > expectedMutationID {
				err = errs.NewSequenceError(fmt.Errorf("mutation id is too high - next expected mutation is: %d", expectedMutationID))
				stop = true
				return false
			}

			if m.ID < expectedMutationID {
				err = errs.NewIdempotencyError(errors.New("mutation has already been processed"))
				return false
			}

			err = processMutation(m, userID, exec)

			if err != nil && !isPermanent(err) {
				// Transient error - don't commit and stop
				stop = true
				return false
			}

			// No error or permanent error - mark mutation as processed
			internalError = replicache.SetMutationID(exec, req.ClientID, m.ID)
			if internalError != nil {
				stop = true
				return false
			}
			return true
		})

		if internalError != nil || txErr != nil {
			httperr.ServerError(w,
				fmt.Sprintf("ERROR finalizing batch endpoint - original err: %v, internalError: %v, txErr: %v",
					err, internalError, txErr))
			return
		}

		if err != nil {
			log.Printf("ERROR for mutation %d: %v", m.ID, err)
			var msg = "Internal server error"
			if isUserVisible(err) {
				msg = err.Error()
			}
			infos = append(infos, mutationInfo{
				ID:    m.ID,
				Error: msg,
			})
		}

		if stop {
			break
		}
	}

	err = json.NewEncoder(w).Encode(batchResponse{
		MutationInfos: infos,
	})
	if err != nil {
		log.Printf("ERROR: Could not encode response: %v", err)
	}
}

func processMutation(m mutation, userID int, exec db.ExecFunc) error {
	if m.Name == "" {
		return errs.NewBadRequestError("mutation name is required")
	}

	r := bytes.NewReader(m.Args)
	var err error
	switch m.Name {
	case "createList":
		err = list.Create(r, exec, userID)
	case "deleteList":
		err = list.Delete(r, exec, userID)
	case "createTodo":
		err = todo.Create(r, exec, userID)
	case "updateTodo":
		err = todo.Update(r, exec, userID)
	case "deleteTodo":
		err = todo.Delete(r, exec, userID)
	default:
		return errs.NewBadRequestError("unknown mutation name")
	}
	if err != nil {
		return fmt.Errorf("%s: %w", m.Name, err)
	}
	return nil
}

// isPermanent classifies an error as a either permanent (it is incorrect and will never change no matter
// how many times it is retried) or temporary (it might succeed next time).
func isPermanent(err error) bool {
	var br errs.BadRequestError
	var ua errs.UnauthorizedError
	var id errs.IdempotencyError
	if errors.As(err, &br) || errors.As(err, &ua) || errors.As(err, &id) {
		return true
	}
	return false
}

func isUserVisible(err error) bool {
	var se errs.SequenceError
	if isPermanent(err) || errors.As(err, &se) {
		return true
	}
	return false
}
