package list

import (
	"roci.dev/replicache-sample-todo/serve/db"
)

type List struct {
	ID          int
	OwnerUserID int
}

func Get(d *db.DB, id int) (List, bool, error) {
	output, err := d.Exec("SELECT (OwnerUserId) FROM TodoList WHERE Id = :id", db.Params{"id": id})
	if err != nil {
		return List{}, false, err
	}
	if len(output.Records) == 0 {
		return List{}, false, nil
	}
	return List{
		ID:          id,
		OwnerUserID: int(*output.Records[0][0].LongValue),
	}, true, nil
}

func Create(d *db.DB, list List) error {
	_, err := d.Exec("INSERT INTO TodoList (Id, OwnerUserId) VALUES (:id, :owner)", db.Params{"id": list.ID, "owner": list.OwnerUserID})
	return err
}
