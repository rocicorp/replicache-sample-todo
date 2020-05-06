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
	_, err := db.Exec("DROP DATABASE IF EXISTS test", nil)
	assert.NoError(err)
	err = schema.Create(db, "test")
	assert.NoError(err)
	db.Use("test")

	userID, err := user.Create(db, "foo@foo.com")
	assert.NoError(err)

	err = list.Create(db, list.List{
		ID:          43,
		OwnerUserID: userID,
	})
	assert.NoError(err)

	has, err := Has(db, 44)
	assert.NoError(err)
	assert.False(has)

	exp := Todo{
		ID:       44,
		ListID:   43,
		Text:     "Take out the trash",
		Complete: true,
		Order:    0.5,
	}
	err = Create(db, exp)
	assert.NoError(err)

	has, err = Has(db, 44)
	assert.NoError(err)
	assert.True(has)
}

func TestUpdate(t *testing.T) {
	assert := assert.New(t)
	db := db.New()
	_, err := db.Exec("DROP DATABASE IF EXISTS test", nil)
	assert.NoError(err)
	err = schema.Create(db, "test")
	assert.NoError(err)
	db.Use("test")

	userID, err := user.Create(db, "foo@foo.com")
	assert.NoError(err)

	err = list.Create(db, list.List{
		ID:          1,
		OwnerUserID: userID,
	})
	assert.NoError(err)

	t1 := Todo{
		ID:       1,
		ListID:   1,
		Text:     "text",
		Complete: true,
		Order:    0.1,
	}
	t2 := Todo{
		ID:       2,
		ListID:   1,
		Text:     "text",
		Complete: true,
		Order:    0.1,
	}
	err = Create(db, t1)
	assert.NoError(err)

	err = Create(db, t2)
	assert.NoError(err)

	pb := func(v bool) *bool {
		return &v
	}
	pf := func(v float64) *float64 {
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
		order    *float64
		text     *string
	}{
		{nil, nil, nil},
		{pb(false), nil, nil},
		{nil, pf(0.2), nil},
		{nil, nil, ps("foo")},
		{pb(true), pf(0.3), nil},
		{nil, pf(0.4), ps("bar")},
		{pb(false), pf(0.5), ps("baz")},
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
		err := Update(db, 1, t.complete, t.order, t.text)
		assert.NoError(err, msg)

		act, has, err := Get(db, 1)
		assert.NoError(err, msg)
		assert.True(has, msg)
		assert.Equal(exp, act)
	}
}
