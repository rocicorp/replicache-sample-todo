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

func MarkComplete(d *db.DB, id int, complete bool) error {
	_, err := d.Exec(
		`UPDATE Todo SET Complete=:complete WHERE Id=:id`,
		db.Params{
			"id":       id,
			"complete": complete})
	return err
}

func Has(d *db.DB, id int) (bool, error) {
	output, err := d.Exec("SELECT 1 FROM Todo WHERE Id = :id", db.Params{"id": id})
	if err != nil {
		return false, err
	}
	return len(output.Records) == 1, nil
}

func Get(d *db.DB, id int, ownerUserID int) (t Todo, ok bool, err error) {
	output, err := d.Exec("SELECT t.Id, t.TodoListId, t.Title, t.Complete, t.SortOrder FROM Todo t, TodoList l WHERE t.TodoListId = l.Id AND t.Id = :id AND l.OwnerUserId = :owneruserid",
		db.Params{"id": id, "owneruserid": ownerUserID})
	if err != nil {
		return Todo{}, false, err
	}
	if len(output.Records) == 0 {
		return Todo{}, false, nil
	}

	return Todo{
		ID:       int(*output.Records[0][0].LongValue),
		ListID:   int(*output.Records[0][1].LongValue),
		Text:     *output.Records[0][2].StringValue,
		Complete: *output.Records[0][3].BooleanValue,
		Order:    *output.Records[0][4].DoubleValue,
	}, true, nil
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
