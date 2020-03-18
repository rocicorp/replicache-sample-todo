package user

import (
	"roci.dev/replicache-sample-todo/serve/db"
)

func Create(d *db.DB, id int) error {
	_, err := d.Exec("INSERT INTO User (Id) VALUES (:id)", db.Params{"id": id})
	if err != nil {
		return err
	}
	return nil
}

func Has(d *db.DB, id int) (bool, error) {
	output, err := d.Exec("SELECT 1 FROM User WHERE Id = :id", db.Params{"id": id})
	if err != nil {
		return false, err
	}
	return len(output.Records) == 1, nil
}
