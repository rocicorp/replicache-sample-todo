package todo

import (
	"roci.dev/replicache-sample-todo/serve/db"
)

type Todo struct {
	ID       int
	ListID   int
	Text     string
	Complete bool
	Order    float64
}

func Create(d *db.DB, todo Todo) error {
	_, err := d.Exec(
		`INSERT INTO Todo (Id, TodoListId, Title, Complete, SortOrder)
		VALUES (:id, :listid, :title, :complete, :order)`,
		db.Params{
			"id":       todo.ID,
			"listid":   todo.ListID,
			"title":    todo.Text,
			"complete": todo.Complete,
			"order":    todo.Order})
	return err
}

func Has(d *db.DB, id int) (bool, error) {
	output, err := d.Exec("SELECT 1 FROM Todo WHERE Id = :id", db.Params{"id": id})
	if err != nil {
		return false, err
	}
	return len(output.Records) == 1, nil
}
