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
	_, err := db.Exec("DROP DATABASE IF EXISTS test")
	assert.NoError(err)
	err = schema.Create(db, "test")
	assert.NoError(err)
	db.Use("test")

	has, err := Has(db, 42)
	assert.NoError(err)
	assert.False(has)
	err = Create(db, 42)
	assert.NoError(err)
	has, err = Has(db, 42)
	assert.NoError(err)
	assert.True(has)
}
