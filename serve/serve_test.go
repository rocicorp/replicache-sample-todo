package serve

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"roci.dev/replicache-sample-todo/serve/db"
	"roci.dev/replicache-sample-todo/serve/model/schema"
	"roci.dev/replicache-sample-todo/serve/model/user"
)

func TestAuthenticate(t *testing.T) {
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
		authorizationHeader string
		wantCode            int
		wantUserID          int
	}{
		{strconv.Itoa(userID), http.StatusOK, userID},
		{"0", http.StatusBadRequest, 0},
		{"abc", http.StatusBadRequest, 0},
		{"-1", http.StatusBadRequest, 0},
		{"", http.StatusUnauthorized, 0},
		{"111111111", http.StatusUnauthorized, 0},
		{"99999999999999999999", http.StatusBadRequest, 0},
	}

	for _, t := range tc {
		r := httptest.NewRequest("", "/", nil)
		r.Header.Set("Authorization", t.authorizationHeader)
		w := httptest.NewRecorder()
		userID := authenticate(db, w, r)
		assert.Equal(t.wantUserID, userID)
		assert.Equal(t.wantCode, w.Result().StatusCode)
	}
}
