package user

import (
	"roci.dev/replicache-sample-todo/serve/db"
)

func Create(exec db.ExecFunc, email string) (int, error) {
	_, err := exec("INSERT INTO User (Email) VALUES (:email)", db.Params{"email": email})
	if err != nil {
		return 0, err
	}
	// Blech, there's probably some better way to do this, but whatevs.
	out, err := exec("SELECT LAST_INSERT_ID()", db.Params{})
	if err != nil {
		return 0, err
	}
	return int(*out.Records[0][0].LongValue), nil
}

func FindByEmail(exec db.ExecFunc, email string) (int, error) {
	output, err := exec("SELECT Id FROM User WHERE Email = :email", db.Params{"email": email})
	if err != nil {
		return 0, err
	}
	if len(output.Records) == 0 {
		return 0, nil
	}
	return int(*output.Records[0][0].LongValue), nil
}

func Has(exec db.ExecFunc, id int) (bool, error) {
	output, err := exec("SELECT 1 FROM User WHERE Id = :id", db.Params{"id": id})
	if err != nil {
		return false, err
	}
	return len(output.Records) == 1, nil
}
