package todo

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/model/list"
	"roci.dev/replicache-sample-todo/serve/model/schema"
	"roci.dev/replicache-sample-todo/serve/model/user"
)

func TestBasic(t *testing.T) {
	assert := assert.New(t)
	db := db.New()
	_, err := db.ExecStatement("DROP DATABASE IF EXISTS test", nil)
	assert.NoError(err)
	err = schema.Create(db, "test")
	assert.NoError(err)
	db.Use("test")

	userID, err := user.Create(db.ExecStatement, "foo@foo.com")
	assert.NoError(err)

	err = list.Create(db.ExecStatement, list.List{
		ID:          43,
		OwnerUserID: userID,
	})
	assert.NoError(err)

	has, err := Has(db.ExecStatement, 44)
	assert.NoError(err)
	assert.False(has)

	exp := Todo{
		ID:       44,
		ListID:   43,
		Text:     "Take out the trash",
		Complete: true,
		Order:    "e",
	}
	err = Create(db.ExecStatement, exp)
	assert.NoError(err)

	has, err = Has(db.ExecStatement, 44)
	assert.NoError(err)
	assert.True(has)
}

func TestUpdate(t *testing.T) {
	assert := assert.New(t)
	db := db.New()
	_, err := db.ExecStatement("DROP DATABASE IF EXISTS test", nil)
	assert.NoError(err)
	err = schema.Create(db, "test")
	assert.NoError(err)
	db.Use("test")

	t1 := Todo{
		ID:       1,
		ListID:   1,
		Text:     "text",
		Complete: true,
		Order:    "a",
	}
	t2 := Todo{
		ID:       2,
		ListID:   1,
		Text:     "text",
		Complete: true,
		Order:    "a",
	}
	err = Create(db.ExecStatement, t1)
	assert.NoError(err)

	err = Create(db.ExecStatement, t2)
	assert.NoError(err)

	pb := func(v bool) *bool {
		return &v
	}
	ps := func(v string) *string {
		return &v
	}

	exp := OwnedTodo{
		Todo:        t1,
		OwnerUserID: 1,
	}
	tc := []struct {
		complete *bool
		order    *string
		text     *string
	}{
		{nil, nil, nil},
		{pb(false), nil, nil},
		{nil, ps("b"), nil},
		{nil, nil, ps("foo")},
		{pb(true), ps("c"), nil},
		{nil, ps("d"), ps("bar")},
		{pb(false), ps("e"), ps("baz")},
	}

	for i, t := range tc {
		msg := fmt.Sprintf("test case %d", i)
		if t.complete != nil {
			exp.Complete = *t.complete
		}
		if t.order != nil {
			exp.Order = *t.order
		}
		if t.text != nil {
			exp.Text = *t.text
		}
		err := Update(db.ExecStatement, 1, t.complete, t.order, t.text)
		assert.NoError(err, msg)

		act, has, err := Get(db.ExecStatement, 1)
		assert.NoError(err, msg)
		assert.True(has, msg)
		assert.Equal(exp, act)
	}
}

func TestDelete(t *testing.T) {
	assert := assert.New(t)
	db := db.New()
	_, err := db.ExecStatement("DROP DATABASE IF EXISTS test", nil)
	assert.NoError(err)
	err = schema.Create(db, "test")
	assert.NoError(err)
	db.Use("test")

	t1 := Todo{
		ID:       1,
		ListID:   1,
		Text:     "text",
		Complete: true,
		Order:    "a",
	}
	err = Create(db.ExecStatement, t1)
	assert.NoError(err)

	t2 := Todo{
		ID:       2,
		ListID:   1,
		Text:     "text",
		Complete: true,
		Order:    "a",
	}
	err = Create(db.ExecStatement, t2)
	assert.NoError(err)

	f := func(id int, wantT1, wantT2 bool) {
		err := Delete(db.ExecStatement, id)
		assert.NoError(err)

		hasT1, err := Has(db.ExecStatement, 1)
		assert.NoError(err)
		hasT2, err := Has(db.ExecStatement, 2)
		assert.NoError(err)

		assert.Equal(wantT1, hasT1)
		assert.Equal(wantT2, hasT2)
	}

	f(3, true, true)
	f(2, true, false)
	f(2, true, false)
	f(1, false, false)
}
