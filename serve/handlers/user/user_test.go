package user

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
	"roci.dev/replicache-sample-todo/serve/model/schema"
)

func TestLogin(t *testing.T) {
	assert := assert.New(t)

	db := db.New()
	_, err := db.Exec("DROP DATABASE IF EXISTS test", nil)
	assert.NoError(err)
	err = schema.Create(db, "test")
	assert.NoError(err)

	db.Use("test")

	tc := []struct {
		request      string
		wantCode     int
		wantUserID   int
		wantResponse string
		wantNewList  bool
	}{
		{``, http.StatusBadRequest, 0, "EOF", false},
		{`{}`, http.StatusBadRequest, 0, "email field is required", false},
		{`{"email":"foo@foo.com"}`, http.StatusOK, 1, "", true},
		{`{"email":"foo@bar.com"}`, http.StatusOK, 2, "", true},
		{`{"email":"foo@foo.com"}`, http.StatusOK, 1, "", false},
	}

	for i, t := range tc {
		msg := fmt.Sprintf("test case %d", i)
		w := httptest.NewRecorder()
		prevMax, err := list.GetMax(db)
		assert.NoError(err)
		Login(w, httptest.NewRequest("POST", "/serve/login", strings.NewReader(t.request)), db)
		assert.Equal(t.wantCode, w.Result().StatusCode)
		body := &bytes.Buffer{}
		_, err = io.Copy(body, w.Result().Body)
		assert.NoError(err)
		if t.wantUserID > 0 {
			assert.Equal(fmt.Sprintf(`{"id":%d}`, t.wantUserID)+"\n", string(body.Bytes()))
		}
		if t.wantResponse != "" {
			assert.Equal(t.wantResponse, string(body.Bytes()))
		}
		currentMax, err := list.GetMax(db)
		assert.NoError(err, msg)
		if t.wantNewList {
			assert.Equal(prevMax+1, currentMax, msg)
			newList, has, err := list.Get(db, currentMax)
			assert.NoError(err, msg)
			assert.True(has, msg)
			assert.Equal(t.wantUserID, newList.OwnerUserID)
		} else {
			assert.Equal(prevMax, currentMax)
		}
	}
}
