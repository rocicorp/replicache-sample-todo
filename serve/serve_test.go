package serve

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
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
	_, err := db.ExecStatement("DROP DATABASE IF EXISTS test", nil)
	assert.NoError(err)
	err = schema.Create(db, "test")
	assert.NoError(err)

	db.Use("test")

	userID, err := user.Create(db.ExecStatement, "foo@foo.com")
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

type doer struct {
	gotTopic string
	gotEvent string
	gotData  interface{}
}

func (d *doer) Do(topic string, event string, data interface{}) error {
	d.gotTopic = topic
	d.gotEvent = event
	d.gotData = data
	return nil
}
func TestPoke(t *testing.T) {
	assert := assert.New(t)

	db := db.New()
	_, err := db.ExecStatement("DROP DATABASE IF EXISTS test", nil)
	assert.NoError(err)
	err = schema.Create(db, "test")
	assert.NoError(err)

	db.Use("test")

	userID, err := user.Create(db.ExecStatement, "foo@foo.com")
	assert.NoError(err)

	os.Setenv("FCM_SERVER_KEY", "test_server_key")

	type tc struct {
		path     string
		wantPoke bool
		// body doesn't matter because serve is lazy and pokes clients
		// just for receiving a request, even if nothing is changed.
	}

	f := func(tt tc) {
		r, err := http.NewRequest("POST", tt.path, &bytes.Buffer{})
		r.Header.Add("Authorization", fmt.Sprintf("%d", userID))
		assert.NoError(err)

		w := httptest.NewRecorder()
		d := &doer{}
		impl(w, r, db, d)

		if !tt.wantPoke {
			assert.Empty(d.gotTopic)
			assert.Empty(d.gotEvent)
			assert.Empty(d.gotData)
			return
		}

		assert.Equal(fmt.Sprintf("u-%d", userID), d.gotTopic)
	}

	f(tc{path: "/serve/replicache-batch", wantPoke: true})
	f(tc{path: "/serve/replicache-client-view", wantPoke: false})
	f(tc{path: "/serve/list-create", wantPoke: true})
	f(tc{path: "/serve/list-delete", wantPoke: true})
	f(tc{path: "/serve/todo-create", wantPoke: true})
	f(tc{path: "/serve/todo-update", wantPoke: true})
	f(tc{path: "/serve/todo-delete", wantPoke: true})
}
