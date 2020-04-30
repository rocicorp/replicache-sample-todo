package todo

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
	"roci.dev/replicache-sample-todo/serve/model/schema"
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

	tc := []struct {
		userID       int
		request      string
		wantCode     int
		wantResponse string
	}{
		{userID, ``, http.StatusBadRequest, `EOF`},
		{userID, `notjson`, http.StatusBadRequest, `invalid character`},
		{userID, `{}`, http.StatusBadRequest, `id field is required`},
		{userID, `{"id":1}`, http.StatusBadRequest, `listID field is required`},
		{userID, `{"id":1,"listID":2}`, http.StatusOK, ``},
		{userID, `{"id":1,"listID":2}`, http.StatusBadRequest, `Specified todo already exists`},
		{7, `{"id":1,"listID":2}`, http.StatusUnauthorized, `Cannot access specified list`},
	}

	for i, t := range tc {
		msg := fmt.Sprintf("test case %d", i)
		w := httptest.NewRecorder()
		Handle(w, httptest.NewRequest("POST", "/serve/todo", strings.NewReader(t.request)), db, t.userID)
		assert.Equal(t.wantCode, w.Result().StatusCode, msg)
		body := &bytes.Buffer{}
		_, err := io.Copy(body, w.Result().Body)
		assert.NoError(err, msg)
		assert.Regexp(t.wantResponse, string(body.Bytes()), msg)
	}
}
