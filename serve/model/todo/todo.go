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

func GetByUser(d *db.DB, userID int) (r []Todo, err error) {
	output, err := d.Exec("SELECT t.Id, t.TodoListId, t.Title, t.Complete, t.SortOrder FROM Todo t, TodoList l WHERE t.TodoListId = l.Id AND l.OwnerUserId = :userid", db.Params{"userid": userID})
	if err != nil {
		return nil, err
	}
	for _, rec := range output.Records {
		r = append(r, Todo{
			ID:       int(*rec[0].LongValue),
			ListID:   int(*rec[1].LongValue),
			Text:     *rec[2].StringValue,
			Complete: *rec[3].BooleanValue,
			Order:    *rec[4].DoubleValue,
		})
	}
	return r, nil
}
