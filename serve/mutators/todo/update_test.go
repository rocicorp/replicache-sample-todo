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

func TestUpdate(t *testing.T) {
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

	tc := []struct {
		label        string
		req          string
		wantErr      error
		wantComplete bool
		wantOrder    float64
		wantText     string
	}{
		{
			label:   "empty-req",
			req:     ``,
			wantErr: errs.NewBadRequestError(`EOF`),
		},
		{
			label:   "not-json",
			req:     `notjson`,
			wantErr: errs.NewBadRequestError(`invalid character 'o' in literal null (expecting 'u')`),
		},
		{
			label:   "no-id",
			req:     `{}`,
			wantErr: errs.NewBadRequestError(`id field is required`),
		},
		{
			label:   "unauth-todo",
			req:     `{"id":2}`,
			wantErr: errs.NewUnauthorizedError(`access unauthorized`),
		},
		{
			label:   "unknown-todo",
			req:     `{"id":3}`,
			wantErr: errs.NewBadRequestError(`specified todo not found`),
		},
		{
			label:        "no-change",
			req:          `{"id":1}`,
			wantErr:      nil,
			wantComplete: false,
			wantOrder:    0.5,
			wantText:     "text",
		},
		{
			label:        "change-complete",
			req:          `{"id":1,"complete":true}`,
			wantErr:      nil,
			wantComplete: true,
			wantOrder:    0.5,
			wantText:     "text",
		},
		{
			label:        "change-complete-nop",
			req:          `{"id":1,"complete":true}`,
			wantErr:      nil,
			wantComplete: true,
			wantOrder:    0.5,
			wantText:     "text",
		},
		{
			label:        "change-complete-zero",
			req:          `{"id":1,"complete":false}`,
			wantErr:      nil,
			wantComplete: false,
			wantOrder:    0.5,
			wantText:     "text",
		},
		{
			label:        "change-order",
			req:          `{"id":1,"order":1.0}`,
			wantErr:      nil,
			wantComplete: false,
			wantOrder:    1.0,
			wantText:     "text",
		},
		{
			label:        "change-text",
			req:          `{"id":1,"text":"bonk"}`,
			wantErr:      nil,
			wantComplete: false,
			wantOrder:    1.0,
			wantText:     "bonk",
		},
		{
			label:        "change-two",
			req:          `{"id":1,"complete":true,"order":0.0}`,
			wantErr:      nil,
			wantComplete: true,
			wantOrder:    0.0,
			wantText:     "bonk",
		},
		{
			label:        "change-three-zero",
			req:          `{"id":1,"complete":false,"order":0.0,"text":""}`,
			wantErr:      nil,
			wantComplete: false,
			wantOrder:    0.0,
			wantText:     "",
		},
	}

	for _, t := range tc {
		err = Update(strings.NewReader(t.req), db.Exec, 1)
		if t.wantErr == nil {
			assert.NoError(err, t.label)
		} else {
			assert.Equal(t.wantErr, err, t.label)
			continue
		}

		got, has, err := todo.Get(db.Exec, 1)
		assert.NoError(err, t.label)
		assert.True(has, t.label)
		assert.Equal(t.wantComplete, got.Complete, t.label)
		assert.Equal(t.wantOrder, got.Order, t.label)
		assert.Equal(t.wantText, got.Text, t.label)
	}

	got, has, err := todo.Get(db.Exec, 2)
	assert.NoError(err)
	assert.True(has)
	want := todo.OwnedTodo{
		Todo:        tt2,
		OwnerUserID: 2,
	}
	assert.Equal(want, got)
}
