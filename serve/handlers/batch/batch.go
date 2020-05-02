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
func Handle(w http.ResponseWriter, r *http.Request, db *db.DB, userID int) {
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
	for _, m := range req.Mutations {
		var err error
		_, txErr := db.Transact(func() bool {
			var currentMutationID, expectedMutationID int64

			currentMutationID, err = replicache.GetMutationID(db, req.ClientID)
			if err != nil {
				return false
			}

			expectedMutationID = currentMutationID + 1
			if m.ID < expectedMutationID {
				err = errs.NewIdempotencyError(errors.New("mutation has already been processed"))
				return false
			}

			if m.ID > expectedMutationID {
				err = errs.NewSequenceError(fmt.Errorf("mutation id is too high - next expected mutation is: %d", expectedMutationID))
				stop = true
				return false
			}

			err = processMutation(m, userID, db)

			if err != nil && !isPermanent(err) {
				// Transient error - don't commit and stop
				stop = true
				return false
			}

			// No error or permanent error - mark mutation as processed
			smErr := replicache.SetMutationID(db, req.ClientID, m.ID)
			if smErr != nil {
				log.Printf("ERROR: Could not SetMutationID: %v", smErr)
				return false
			}
			return true
		})

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

		if txErr != nil {
			log.Printf("ERROR committing transaction for mutation %d: %v", m.ID, txErr)
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

func processMutation(m mutation, userID int, db *db.DB) error {
	if m.Name == "" {
		return errs.NewBadRequestError("mutation name is required")
	}

	r := bytes.NewReader(m.Args)
	var err error
	switch m.Name {
	case "createTodo":
		err = todo.Create(r, db, userID)
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
