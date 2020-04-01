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
		wantReturn   bool
		wantCode     int
		wantResponse string
	}{
		{``, false, http.StatusBadRequest, "EOF"},
		{`{}`, false, http.StatusBadRequest, "email field is required"},
		{`{"email":"foo@foo.com"}`, true, http.StatusOK, `{"id":1}` + "\n"},
		{`{"email":"foo@bar.com"}`, true, http.StatusOK, `{"id":2}` + "\n"},
		{`{"email":"foo@foo.com"}`, true, http.StatusOK, `{"id":1}` + "\n"},
	}

	for i, t := range tc {
		msg := fmt.Sprintf("test case %d", i)
		w := httptest.NewRecorder()
		ret := Login(w, httptest.NewRequest("POST", "/serve/login", strings.NewReader(t.request)), db)
		assert.Equal(t.wantReturn, ret, msg)
		assert.Equal(t.wantCode, w.Result().StatusCode)
		body := &bytes.Buffer{}
		_, err := io.Copy(body, w.Result().Body)
		assert.NoError(err)
		assert.Equal(t.wantResponse, string(body.Bytes()))
	}
}
