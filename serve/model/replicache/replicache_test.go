package replicache

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/model/schema"
)

func TestPutGet(t *testing.T) {
	assert := assert.New(t)

	lolGo := func(i int) *int {
		return &i
	}

	tc := []struct {
		clientID     string
		put          *int // nil means no record
		wantPutError string
		want         int
	}{
		{"foo", nil, "", 0},
		{"foo", lolGo(0), "unexpected new MutationID - expected 1, got 0", 0},
		{"foo", lolGo(1), "", 1},
		{"foo", lolGo(3), "unexpected new MutationID - expected 2, got 1", 1},
		{"bar", nil, "", 0},
	}

	db := db.New()
	_, err := db.ExecStatement("DROP DATABASE IF EXISTS test", nil)
	assert.NoError(err)
	err = schema.Create(db, "test")
	assert.NoError(err)
	db.Use("test")

	for i, t := range tc {
		msg := fmt.Sprintf("test case %d", i)

		if t.put != nil {
			err = SetMutationID(db.ExecStatement, t.clientID, int64(*t.put))
			if t.wantPutError != "" {
				assert.Error(err, t.wantPutError)
			} else {
				assert.NoError(err, msg)
			}
		}

		actual, err := GetMutationID(db.ExecStatement, t.clientID)
		assert.NoError(err, msg)
		assert.Equal(int64(t.want), actual, msg)
	}
}
