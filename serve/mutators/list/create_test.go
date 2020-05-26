package list

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/model/schema"
	"roci.dev/replicache-sample-todo/serve/model/user"
	"roci.dev/replicache-sample-todo/serve/util/errs"
)

func TestCreate(t *testing.T) {
	assert := assert.New(t)

	db := db.New()
	_, err := db.ExecStatement("DROP DATABASE IF EXISTS test", nil)
	assert.NoError(err)
	err = schema.Create(db, "test")
	assert.NoError(err)

	db.Use("test")

	userID, err := user.Create(db.ExecStatement, "foo@foo.com")
	assert.NoError(err)

	otherUserID, err := user.Create(db.ExecStatement, "bar@bar.com")
	assert.NoError(err)

	tc := []struct {
		userID  int
		request string
		wantErr error
	}{
		{userID, ``, errs.NewBadRequestError(`EOF`)},
		{userID, `notjson`, errs.NewBadRequestError(`invalid character 'o' in literal null (expecting 'u')`)},
		{userID, `{}`, errs.NewBadRequestError(`id field is required`)},
		{userID, `{"id":1}`, nil},
		{userID, `{"id":1}`, errs.NewBadRequestError(`specified list already exists: 1`)},
		{otherUserID, `{"id":1}`, errs.NewBadRequestError(`specified list already exists: 1`)},
	}

	for i, t := range tc {
		msg := fmt.Sprintf("test case %d", i)
		err = Create(strings.NewReader(t.request), db.ExecStatement, t.userID)
		if t.wantErr == nil {
			assert.NoError(err, msg)
		} else {
			assert.Equal(t.wantErr, err, msg)
		}
	}
}
