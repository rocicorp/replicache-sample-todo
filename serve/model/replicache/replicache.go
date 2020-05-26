package replicache

import (
	"fmt"

	"roci.dev/replicache-sample-todo/serve/db"
)

// GetMutationID fetches the last processed mutation ID for a client from the database.
// Returns zero and nil if the specified client is new and has no last mutationID.
func GetMutationID(exec db.ExecFunc, clientID string) (int64, error) {
	output, err := exec("SELECT MutationID FROM Replicache WHERE ClientID = :clientid", db.Params{"clientid": clientID})
	if err != nil {
		return 0, err
	}

	if len(output.Records) == 0 {
		return 0, nil
	}

	if len(output.Records) > 1 {
		return int64(0), fmt.Errorf("unexpected number of MutationID records")
	}

	return *output.Records[0][0].LongValue, nil
}

// SetMutationID updates the database with the last processed clientID for a client.
func SetMutationID(exec db.ExecFunc, clientID string, mutationID int64) error {
	last, err := GetMutationID(exec, clientID)
	if err != nil {
		return err
	}
	expected := last + 1
	if mutationID != expected {
		return fmt.Errorf("unexpected new MutationID. Expected %d, got %d", expected, mutationID)
	}
	if last == 0 {
		_, err = exec("INSERT INTO Replicache (ClientID, MutationID) VALUES (:clientid, :mutationid)",
			db.Params{"clientid": clientID, "mutationid": mutationID})
		return err
	}
	_, err = exec("UPDATE Replicache SET MutationID=:mutationid WHERE ClientID=:clientid",
		db.Params{"clientid": clientID, "mutationid": mutationID})
	return err
}
