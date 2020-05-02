package todo

import (
	"fmt"
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

func TestMarkComplete(t *testing.T) {
	assert := assert.New(t)

	db := db.New()
	_, err := db.Exec("DROP DATABASE IF EXISTS test", nil)
	assert.NoError(err)
	err = schema.Create(db, "test")
	assert.NoError(err)

	db.Use("test")

	userID, err := user.Create(db, "foo@foo.com")
	assert.NoError(err)

	otherUserID, err := user.Create(db, "bar@bar.com")
	assert.NoError(err)

	l := list.List{
		ID:          1,
		OwnerUserID: userID,
	}
	err = list.Create(db, l)
	assert.NoError(err)

	l = list.List{
		ID:          2,
		OwnerUserID: otherUserID,
	}
	err = list.Create(db, l)
	assert.NoError(err)

	tt1 := todo.Todo{
		ID:       1,
		ListID:   1,
		Text:     "text",
		Complete: false,
		Order:    0.5,
	}
	err = todo.Create(db, tt1)
	assert.NoError(err)

	tt2 := todo.Todo{
		ID:       2,
		ListID:   2,
		Text:     "text",
		Complete: false,
		Order:    0.5,
	}
	err = todo.Create(db, tt2)
	assert.NoError(err)

	tc := []struct {
		request      string
		wantErr      error
		wantComplete bool
	}{
		{``, errs.NewBadRequestError(`EOF`), false},
		{`notjson`, errs.NewBadRequestError(`invalid character 'o' in literal null (expecting 'u')`), false},
		{`{}`, errs.NewBadRequestError(`id field is required`), false},
		{`{"id":2}`, errs.NewBadRequestError(`specified todo not found: 2`), false},
		{`{"id":3}`, errs.NewBadRequestError(`specified todo not found: 3`), false},
		{`{"id":1}`, nil, false},
		{`{"id":1,"complete":true}`, nil, true},
		{`{"id":1,"complete":true}`, nil, true},
		{`{"id":1,"complete":false}`, nil, false},
	}

	for i, t := range tc {
		msg := fmt.Sprintf("test case %d", i)
		err = MarkComplete(strings.NewReader(t.request), db, 1)
		if t.wantErr == nil {
			assert.NoError(err, msg)
		} else {
			assert.Equal(t.wantErr, err, msg)
		}
		got, has, err := todo.Get(db, 1, 1)
		assert.NoError(err, msg)
		assert.True(has, msg)
		assert.Equal(t.wantComplete, got.Complete)
	}
}
