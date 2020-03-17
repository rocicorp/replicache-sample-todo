package user

import (
	"fmt"

	"roci.dev/replicache-sample-todo/serve/db"
)

func Create(db *db.DB, id int) error {
	_, err := db.Exec(fmt.Sprintf("INSERT INTO User (Id) VALUES (%d)", id))
	if err != nil {
		return err
	}
	return nil
}

func Has(db *db.DB, id int) (bool, error) {
	output, err := db.Exec(fmt.Sprintf("SELECT 1 FROM User WHERE Id = %d", id))
	if err != nil {
		return false, err
	}
	return len(output.Records) == 1, nil
}
