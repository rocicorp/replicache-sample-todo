package todo

import (
	"log"

	"roci.dev/replicache-sample-todo/serve/db"
)

type Todo struct {
	ID          int    `json:"id"`
	ListID      int    `json:"listId"`
	Text        string `json:"text"`
	Complete    bool   `json:"complete"`
	LegacyOrder string `json:"order"`
	Order       string `json:"order_str"`
}

type OwnedTodo struct {
	Todo
	OwnerUserID int `json:"ownerUserID"`
}

func Create(exec db.ExecFunc, todo Todo) error {
	_, err := exec(
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

func Update(exec db.ExecFunc, id int, complete *bool, order *string, title *string) error {
	log.Printf("Updating: id: %#v, complete: %#v, order: %#v, title: %#v", id, complete, order, title)
	_, err := exec(
		`UPDATE Todo SET Complete=COALESCE(:complete,Complete), SortOrder=COALESCE(:order,SortOrder), Title=COALESCE(:title,Title) WHERE Id=:id`,
		db.Params{
			"id":       id,
			"complete": complete,
			"order":    order,
			"title":    title,
		})
	return err
}

func Delete(exec db.ExecFunc, id int) error {
	_, err := exec(`DELETE FROM Todo WHERE Id = :id`, db.Params{"id": id})
	return err
}

func Has(exec db.ExecFunc, id int) (bool, error) {
	output, err := exec("SELECT 1 FROM Todo WHERE Id = :id", db.Params{"id": id})
	if err != nil {
		return false, err
	}
	return len(output.Records) == 1, nil
}

func Get(exec db.ExecFunc, id int) (t OwnedTodo, ok bool, err error) {
	output, err := exec("SELECT t.Id, t.TodoListId, t.Title, t.Complete, t.SortOrder, l.OwnerUserID FROM Todo t, TodoList l WHERE t.TodoListId = l.Id AND t.Id = :id",
		db.Params{"id": id})
	if err != nil {
		return OwnedTodo{}, false, err
	}
	if len(output.Records) == 0 {
		return OwnedTodo{}, false, nil
	}

	return OwnedTodo{
		Todo: Todo{
			ID:       int(*output.Records[0][0].LongValue),
			ListID:   int(*output.Records[0][1].LongValue),
			Text:     *output.Records[0][2].StringValue,
			Complete: *output.Records[0][3].BooleanValue,
			Order:    *output.Records[0][4].StringValue,
		},
		OwnerUserID: int(*output.Records[0][5].LongValue),
	}, true, nil
}

func GetByUser(exec db.ExecFunc, userID int) (r []Todo, err error) {
	output, err := exec("SELECT t.Id, t.TodoListId, t.Title, t.Complete, t.SortOrder FROM Todo t, TodoList l WHERE t.TodoListId = l.Id AND l.OwnerUserId = :userid", db.Params{"userid": userID})
	if err != nil {
		return nil, err
	}
	for _, rec := range output.Records {
		r = append(r, Todo{
			ID:       int(*rec[0].LongValue),
			ListID:   int(*rec[1].LongValue),
			Text:     *rec[2].StringValue,
			Complete: *rec[3].BooleanValue,
			Order:    *rec[4].StringValue,
		})
	}
	return r, nil
}
