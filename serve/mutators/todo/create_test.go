package todo

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/model/list"
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

	otherUserID, err := user.Create(db.ExecStatement, "bar@bar.com")
	assert.NoError(err)

	l := list.List{
		ID:          3,
		OwnerUserID: otherUserID,
	}
	err = list.Create(db.ExecStatement, l)
	assert.NoError(err)

	tc := []struct {
		userID  int
		request string
		wantErr error
	}{
		{1, ``, errs.NewBadRequestError(`EOF`)},
		{1, `notjson`, errs.NewBadRequestError(`invalid character 'o' in literal null (expecting 'u')`)},
		{1, `{}`, errs.NewBadRequestError(`id field is required`)},
		{1, `{"id":1}`, errs.NewBadRequestError(`listID field is required`)},
		{1, `{"id":1,"listID":1}`, nil},
		{1, `{"id":1,"listID":1}`, errs.NewBadRequestError(`specified todo already exists: 1`)},
		{1, `{"id":2,"listID":7}`, errs.NewBadRequestError(`specified list does not exist: 7`)},
		{1, `{"id":2,"listID":3}`, errs.NewUnauthorizedError(`cannot access specified list: 3`)},
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
