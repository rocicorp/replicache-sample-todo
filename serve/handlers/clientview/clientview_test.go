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
	"roci.dev/replicache-sample-todo/serve/model/replicache"
	"roci.dev/replicache-sample-todo/serve/model/schema"
	"roci.dev/replicache-sample-todo/serve/model/todo"
)

func TestClientView(t *testing.T) {
	assert := assert.New(t)

	db := db.New()
	_, err := db.ExecStatement("DROP DATABASE IF EXISTS test", nil)
	assert.NoError(err)
	err = schema.Create(db, "test")
	assert.NoError(err)

	db.Use("test")

	err = todo.Create(db.ExecStatement, todo.Todo{
		ID:     3,
		ListID: 1,
		Order:  "a0",
	})
	assert.NoError(err)

	err = replicache.SetMutationID(db.ExecStatement, "c1", int64(1))
	assert.NoError(err)

	tc := []struct {
		userID       int
		req          string
		wantReturn   bool
		wantCode     int
		wantResponse string
	}{
		{1, `{"clientID":"c1"}`, true, http.StatusOK, `{"clientView":{"/list/1":{"id":1,"ownerUserID":1},"/todo/3":{"id":3,"listId":1,"text":"","complete":false,"order":"a0"}},"lastMutationID":1}`},
		{1, `{"clientID":"c2"}`, true, http.StatusOK, `{"clientView":{"/list/1":{"id":1,"ownerUserID":1},"/todo/3":{"id":3,"listId":1,"text":"","complete":false,"order":"a0"}},"lastMutationID":0}`},
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
