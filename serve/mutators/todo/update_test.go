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
	_, err := db.ExecStatement("DROP DATABASE IF EXISTS test", nil)
	assert.NoError(err)
	err = schema.Create(db, "test")
	assert.NoError(err)

	db.Use("test")

	otherUserID, err := user.Create(db.ExecStatement, "bar@bar.com")
	assert.NoError(err)

	l := list.List{
		ID:          1,
		OwnerUserID: 1,
	}
	err = list.Create(db.ExecStatement, l)
	assert.NoError(err)

	l = list.List{
		ID:          2,
		OwnerUserID: otherUserID,
	}
	err = list.Create(db.ExecStatement, l)
	assert.NoError(err)

	tt1 := todo.Todo{
		ID:       1,
		ListID:   1,
		Text:     "text",
		Complete: false,
		Order:    "e",
	}
	err = todo.Create(db.ExecStatement, tt1)
	assert.NoError(err)

	tt2 := todo.Todo{
		ID:       2,
		ListID:   2,
		Text:     "text",
		Complete: false,
		Order:    "e",
	}
	err = todo.Create(db.ExecStatement, tt2)
	assert.NoError(err)

	tc := []struct {
		label        string
		req          string
		wantErr      error
		wantComplete bool
		wantOrder    string
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
			wantOrder:    "e",
			wantText:     "text",
		},
		{
			label:        "change-complete",
			req:          `{"id":1,"complete":true}`,
			wantErr:      nil,
			wantComplete: true,
			wantOrder:    "e",
			wantText:     "text",
		},
		{
			label:        "change-complete-nop",
			req:          `{"id":1,"complete":true}`,
			wantErr:      nil,
			wantComplete: true,
			wantOrder:    "e",
			wantText:     "text",
		},
		{
			label:        "change-complete-zero",
			req:          `{"id":1,"complete":false}`,
			wantErr:      nil,
			wantComplete: false,
			wantOrder:    "e",
			wantText:     "text",
		},
		{
			label:        "change-order",
			req:          `{"id":1,"order":"z"}`,
			wantErr:      nil,
			wantComplete: false,
			wantOrder:    "z",
			wantText:     "text",
		},
		{
			label:        "change-text",
			req:          `{"id":1,"text":"bonk"}`,
			wantErr:      nil,
			wantComplete: false,
			wantOrder:    "z",
			wantText:     "bonk",
		},
		{
			label:        "change-two",
			req:          `{"id":1,"complete":true,"order":""}`,
			wantErr:      nil,
			wantComplete: true,
			wantOrder:    "",
			wantText:     "bonk",
		},
		{
			label:        "change-three-zero",
			req:          `{"id":1,"complete":false,"order":"","text":""}`,
			wantErr:      nil,
			wantComplete: false,
			wantOrder:    "",
			wantText:     "",
		},
	}

	for _, t := range tc {
		err = Update(strings.NewReader(t.req), db.ExecStatement, 1)
		if t.wantErr == nil {
			assert.NoError(err, t.label)
		} else {
			assert.Equal(t.wantErr, err, t.label)
			continue
		}

		got, has, err := todo.Get(db.ExecStatement, 1)
		assert.NoError(err, t.label)
		assert.True(has, t.label)
		assert.Equal(t.wantComplete, got.Complete, t.label)
		assert.Equal(t.wantOrder, got.Order, t.label)
		assert.Equal(t.wantText, got.Text, t.label)
	}

	got, has, err := todo.Get(db.ExecStatement, 2)
	assert.NoError(err)
	assert.True(has)
	want := todo.OwnedTodo{
		Todo:        tt2,
		OwnerUserID: 2,
	}
	assert.Equal(want, got)
}
