package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"roci.dev/replicache-sample-todo/serve/db"
)

func TestSchema(t *testing.T) {
	assert := assert.New(t)
	db := db.New()
	_, err := db.Exec("DROP DATABASE 'foo'")
	err = Create(db, "foo")
	assert.NoError(err)
	out, err := db.Exec("SHOW DATABASES LIKE 'foo'")
	assert.Equal(1, len(out.Records))
}
