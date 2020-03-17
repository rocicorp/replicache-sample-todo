package todo

import (
	"fmt"

	"roci.dev/replicache-sample-todo/serve/db"
)

type Todo struct {
	ID       int
	ListID   int
	Text     string
	Complete bool
	Order    float64
}

func Create(db *db.DB, todo Todo) error {
	_, err := db.Exec(fmt.Sprintf(
		"INSERT INTO Todo (Id, TodoListId, Title, Complete, SortOrder) VALUES (%d, %d, '%s', %t, %f)",
		todo.ID, todo.ListID, todo.Text, todo.Complete, todo.Order))
	return err
}

func Has(db *db.DB, id int) (bool, error) {
	output, err := db.Exec(fmt.Sprintf("SELECT 1 FROM Todo WHERE Id = %d", id))
	if err != nil {
		return false, err
	}
	return len(output.Records) == 1, nil
}
