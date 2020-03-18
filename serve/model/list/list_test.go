package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"roci.dev/replicache-sample-todo/serve/db"
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

	err = user.Create(db, 43)
	assert.NoError(err)

	act, has, err := Get(db, 42)
	assert.NoError(err)
	assert.False(has)
	assert.Equal(List{}, act)

	exp := List{
		ID:          42,
		OwnerUserID: 43,
	}
	err = Create(db, exp)
	assert.NoError(err)

	act, has, err = Get(db, 42)
	assert.NoError(err)
	assert.True(has)
	assert.Equal(exp, act)

	exp2 := List{
		ID:          43,
		OwnerUserID: 43,
	}
	err = Create(db, exp2)
	assert.NoError(err)

	act2, err := GetByUser(db, 43)
	assert.NoError(err)

	assert.Equal([]List{exp, exp2}, act2)
}
