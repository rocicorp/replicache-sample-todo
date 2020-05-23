package list

import (
	"roci.dev/replicache-sample-todo/serve/db"
)

type List struct {
	ID          int `json:"id"`
	OwnerUserID int `json:"ownerUserID"`
}

func Get(exec db.ExecFunc, id int) (List, bool, error) {
	output, err := exec("SELECT (OwnerUserId) FROM TodoList WHERE Id = :id", db.Params{"id": id})
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

func GetByUser(exec db.ExecFunc, userID int) (r []List, err error) {
	output, err := exec("SELECT Id FROM TodoList WHERE OwnerUserId = :ownerId", db.Params{"ownerId": userID})
	if err != nil {
		return nil, err
	}
	for _, rec := range output.Records {
		r = append(r, List{
			ID:          int(*rec[0].LongValue),
			OwnerUserID: userID,
		})
	}
	return r, nil
}

func GetMax(exec db.ExecFunc) (int, error) {
	out, err := exec("SELECT Max(Id) FROM TodoList", db.Params{})
	if err != nil {
		return 0, err
	}
	if out.Records[0][0].IsNull != nil {
		return 0, nil
	}
	return int(*out.Records[0][0].LongValue), nil
}

func Create(exec db.ExecFunc, list List) error {
	_, err := exec("INSERT INTO TodoList (Id, OwnerUserId) VALUES (:id, :owner)", db.Params{"id": list.ID, "owner": list.OwnerUserID})
	return err
}
