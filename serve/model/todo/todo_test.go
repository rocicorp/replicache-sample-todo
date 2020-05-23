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

	userID, err := user.Create(db.Exec, "foo@foo.com")
	assert.NoError(err)

	err = list.Create(db.Exec, list.List{
		ID:          43,
		OwnerUserID: userID,
	})
	assert.NoError(err)

	has, err := Has(db.Exec, 44)
	assert.NoError(err)
	assert.False(has)

	exp := Todo{
		ID:       44,
		ListID:   43,
		Text:     "Take out the trash",
		Complete: true,
		Order:    0.5,
	}
	err = Create(db.Exec, exp)
	assert.NoError(err)

	has, err = Has(db.Exec, 44)
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

	userID, err := user.Create(db.Exec, "foo@foo.com")
	assert.NoError(err)

	err = list.Create(db.Exec, list.List{
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
	err = Create(db.Exec, t1)
	assert.NoError(err)

	err = Create(db.Exec, t2)
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
		err := Update(db.Exec, 1, t.complete, t.order, t.text)
		assert.NoError(err, msg)

		act, has, err := Get(db.Exec, 1)
		assert.NoError(err, msg)
		assert.True(has, msg)
		assert.Equal(exp, act)
	}
}

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

	err = list.Create(db.Exec, list.List{
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
	err = Create(db.Exec, t1)
	assert.NoError(err)

	t2 := Todo{
		ID:       2,
		ListID:   1,
		Text:     "text",
		Complete: true,
		Order:    0.1,
	}
	err = Create(db.Exec, t2)
	assert.NoError(err)

	f := func(id int, wantT1, wantT2 bool) {
		err := Delete(db.Exec, id)
		assert.NoError(err)

		hasT1, err := Has(db.Exec, 1)
		assert.NoError(err)
		hasT2, err := Has(db.Exec, 2)
		assert.NoError(err)

		assert.Equal(wantT1, hasT1)
		assert.Equal(wantT2, hasT2)
	}

	f(3, true, true)
	f(2, true, false)
	f(2, true, false)
	f(1, false, false)
}
