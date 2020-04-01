package todo

import (
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
