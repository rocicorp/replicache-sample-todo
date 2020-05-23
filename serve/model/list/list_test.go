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

	userID, err := user.Create(db.Exec, "foo@foo.com")
	assert.NoError(err)

	act, has, err := Get(db.Exec, 42)
	assert.NoError(err)
	assert.False(has)
	assert.Equal(List{}, act)

	max, err := GetMax(db.Exec)
	assert.NoError(err)
	assert.Equal(0, max)

	exp := List{
		ID:          42,
		OwnerUserID: userID,
	}
	err = Create(db.Exec, exp)
	assert.NoError(err)

	max, err = GetMax(db.Exec)
	assert.NoError(err)
	assert.Equal(42, max)

	act, has, err = Get(db.Exec, 42)
	assert.NoError(err)
	assert.True(has)
	assert.Equal(exp, act)

	exp2 := List{
		ID:          43,
		OwnerUserID: userID,
	}
	err = Create(db.Exec, exp2)
	assert.NoError(err)

	act2, err := GetByUser(db.Exec, userID)
	assert.NoError(err)

	assert.Equal([]List{exp, exp2}, act2)
}
