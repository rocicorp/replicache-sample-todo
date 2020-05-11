package clientview

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/model/list"
	"roci.dev/replicache-sample-todo/serve/model/replicache"
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

	err = replicache.SetMutationID(db, "c1", int64(1))
	assert.NoError(err)

	tc := []struct {
		userID       int
		req          string
		wantReturn   bool
		wantCode     int
		wantResponse string
	}{
		{userID, `{"clientID":"c1"}`, true, http.StatusOK, `{"clientView":{"/list/2":{"id":2,"ownerUserID":1},"/todo/3":{"id":3,"listId":2,"text":"","complete":false,"order":0}},"lastMutationID":1}`},
		{userID, `{"clientID":"c2"}`, true, http.StatusOK, `{"clientView":{"/list/2":{"id":2,"ownerUserID":1},"/todo/3":{"id":3,"listId":2,"text":"","complete":false,"order":0}},"lastMutationID":0}`},
	}

	for i, t := range tc {
		msg := fmt.Sprintf("test case %d", i)
		w := httptest.NewRecorder()
		Handle(w, httptest.NewRequest("POST", "/serve/clientview", strings.NewReader(t.req)), db, t.userID)
		assert.Equal(t.wantCode, w.Result().StatusCode, msg)
		body := &bytes.Buffer{}
		_, err := io.Copy(body, w.Result().Body)
		assert.NoError(err, msg)
		assert.Equal(t.wantResponse+"\n", string(body.Bytes()), msg)
	}
}
