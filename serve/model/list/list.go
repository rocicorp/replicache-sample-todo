package list

import (
	"fmt"

	"roci.dev/replicache-sample-todo/serve/db"
)

type List struct {
	ID          int
	OwnerUserID int
}

func Get(db *db.DB, id int) (List, bool, error) {
	output, err := db.Exec(fmt.Sprintf("SELECT (OwnerUserId) FROM TodoList WHERE Id = %d", id))
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

func Create(db *db.DB, list List) error {
	_, err := db.Exec(fmt.Sprintf("INSERT INTO TodoList (Id, OwnerUserId) VALUES (%d, %d)", list.ID, list.OwnerUserID))
	return err
}
