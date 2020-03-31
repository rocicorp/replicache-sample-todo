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

	userID, err := user.Create(db, "foo@foo.com")
	assert.NoError(err)

	act, has, err := Get(db, 42)
	assert.NoError(err)
	assert.False(has)
	assert.Equal(List{}, act)

	exp := List{
		ID:          42,
		OwnerUserID: userID,
	}
	err = Create(db, exp)
	assert.NoError(err)

	act, has, err = Get(db, 42)
	assert.NoError(err)
	assert.True(has)
	assert.Equal(exp, act)

	exp2 := List{
		ID:          43,
		OwnerUserID: userID,
	}
	err = Create(db, exp2)
	assert.NoError(err)

	act2, err := GetByUser(db, userID)
	assert.NoError(err)

	assert.Equal([]List{exp, exp2}, act2)
}
