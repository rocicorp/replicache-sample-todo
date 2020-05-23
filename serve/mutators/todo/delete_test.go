package todo

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/model/list"
	"roci.dev/replicache-sample-todo/serve/model/schema"
	"roci.dev/replicache-sample-todo/serve/model/todo"
	"roci.dev/replicache-sample-todo/serve/model/user"
	"roci.dev/replicache-sample-todo/serve/util/errs"
)

func TestDelete(t *testing.T) {
	assert := assert.New(t)

	db := db.New()
	_, err := db.Exec("DROP DATABASE IF EXISTS test", nil)
	assert.NoError(err)
	err = schema.Create(db, "test")
	assert.NoError(err)

	db.Use("test")

	userID, err := user.Create(db.Exec, "foo@foo.com")
	assert.NoError(err)

	otherUserID, err := user.Create(db.Exec, "bar@bar.com")
	assert.NoError(err)

	l := list.List{
		ID:          1,
		OwnerUserID: userID,
	}
	err = list.Create(db.Exec, l)
	assert.NoError(err)

	l = list.List{
		ID:          2,
		OwnerUserID: otherUserID,
	}
	err = list.Create(db.Exec, l)
	assert.NoError(err)

	tt1 := todo.Todo{
		ID:       1,
		ListID:   1,
		Text:     "text",
		Complete: false,
		Order:    0.5,
	}
	err = todo.Create(db.Exec, tt1)
	assert.NoError(err)

	tt2 := todo.Todo{
		ID:       2,
		ListID:   2,
		Text:     "text",
		Complete: false,
		Order:    0.5,
	}
	err = todo.Create(db.Exec, tt2)
	assert.NoError(err)

	f := func(req string, wantErr error, wantTodo1, wantTodo2 bool) {
		err = Delete(strings.NewReader(req), db.Exec, 1)
		if wantErr == nil {
			assert.NoError(err)
		} else {
			assert.Equal(wantErr, err)
		}

		hasT1, err := todo.Has(db.Exec, 1)
		assert.NoError(err)
		assert.Equal(wantTodo1, hasT1)

		hasT2, err := todo.Has(db.Exec, 2)
		assert.NoError(err)
		assert.Equal(wantTodo2, hasT2)
	}

	f(``, errs.NewBadRequestError(`EOF`), true, true)
	f(`notjson`, errs.NewBadRequestError(`invalid character 'o' in literal null (expecting 'u')`), true, true)
	f(`{}`, errs.NewBadRequestError(`id field is required`), true, true)
	f(`{"id":2}`, errs.NewUnauthorizedError(`access unauthorized`), true, true)
	f(`{"id":3}`, errs.NewBadRequestError(`todo not found`), true, true)
	f(`{"id":1}`, nil, false, true)
	f(`{"id":1}`, errs.NewBadRequestError(`todo not found`), false, true)
}
