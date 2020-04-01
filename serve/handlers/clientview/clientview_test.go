package clientview

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/model/list"
	"roci.dev/replicache-sample-todo/serve/model/schema"
	"roci.dev/replicache-sample-todo/serve/model/todo"
	"roci.dev/replicache-sample-todo/serve/model/user"
)

func TestClientView(t *testing.T) {
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
		ID:          2,
		OwnerUserID: userID,
	})
	assert.NoError(err)

	err = todo.Create(db, todo.Todo{
		ID:     3,
		ListID: 2,
	})
	assert.NoError(err)

	tc := []struct {
		userID       int
		wantReturn   bool
		wantCode     int
		wantResponse string
	}{
		{userID, true, http.StatusOK, `{"clientView":{"/list/2":{"id":2,"ownerUserID":1},"/todo/3":{"id":3,"listId":2,"text":"","complete":false,"order":0}}`},
	}

	for i, t := range tc {
		msg := fmt.Sprintf("test case %d", i)
		w := httptest.NewRecorder()
		ret := Handle(w, httptest.NewRequest("POST", "/serve/clientview", nil), db, t.userID)
		assert.Equal(t.wantReturn, ret, msg)
		assert.Equal(t.wantCode, w.Result().StatusCode)
		body := &bytes.Buffer{}
		_, err := io.Copy(body, w.Result().Body)
		assert.NoError(err)
		assert.Regexp(t.wantResponse, string(body.Bytes()))
	}
}
