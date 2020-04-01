package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/model/schema"
)

func TestBasic(t *testing.T) {
	assert := assert.New(t)
	db := db.New()
	_, err := db.Exec("DROP DATABASE IF EXISTS test", nil)
	assert.NoError(err)
	err = schema.Create(db, "test")
	assert.NoError(err)
	db.Use("test")

	id, err := FindByEmail(db, "foo@foo.com")
	assert.NoError(err)
	assert.Equal(0, id)
	id, err = Create(db, "foo@foo.com")
	assert.NoError(err)
	assert.NotEqual(0, id)
	has, err := Has(db, id)
	assert.NoError(err)
	assert.True(has)
	found, err := FindByEmail(db, "foo@foo.com")
	assert.NoError(err)
	assert.Equal(id, found)
}
