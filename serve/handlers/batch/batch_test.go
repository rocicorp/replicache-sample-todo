package batch

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

func TestHandle(t *testing.T) {
	assert := assert.New(t)

	db := db.New()
	_, err := db.Exec("DROP DATABASE IF EXISTS test", nil)
	assert.NoError(err)
	err = schema.Create(db, "test")
	assert.NoError(err)

	db.Use("test")

	userID, err := user.Create(db.Exec, "foo@foo.com")
	assert.NoError(err)

	err = list.Create(db.Exec, list.List{
		ID:          1,
		OwnerUserID: userID,
	})
	assert.NoError(err)

	tc := []struct {
		label          string
		req            string
		userID         int
		wantCode       int
		wantResponse   string
		wantMutationID int64
		wantNumTodos   int
	}{
		{
			label:          "empty-request",
			req:            ``,
			userID:         0,
			wantCode:       http.StatusBadRequest,
			wantResponse:   "EOF",
			wantMutationID: 0,
			wantNumTodos:   0,
		},
		{
			label:          "not-json",
			req:            "not-json",
			userID:         0,
			wantCode:       http.StatusBadRequest,
			wantResponse:   "invalid character 'o' in literal null (expecting 'u')",
			wantMutationID: 0,
			wantNumTodos:   0,
		},
		{
			label:          "missing-clientid",
			req:            `{}`,
			userID:         0,
			wantCode:       http.StatusBadRequest,
			wantResponse:   "clientID is required",
			wantMutationID: 0,
			wantNumTodos:   0,
		},
		{
			label:          "no-mutations",
			req:            `{"clientID":"c1"}`,
			userID:         0,
			wantCode:       http.StatusOK,
			wantResponse:   `{"mutationInfos":[]}` + "\n",
			wantMutationID: 0,
			wantNumTodos:   0,
		},
		{
			label:          "idempotency-0",
			req:            `{"clientID":"c1","mutations":[{"id":0,"name":"createTodo"}]}`,
			userID:         1,
			wantCode:       http.StatusOK,
			wantResponse:   `{"mutationInfos":[{"id":0,"error":"mutation has already been processed"}]}` + "\n",
			wantMutationID: 0,
			wantNumTodos:   0,
		},
		{
			label:          "timetravel", // TODO: others
			req:            `{"clientID":"c1","mutations":[{"id":2,"name":"createTodo"}]}`,
			userID:         1,
			wantCode:       http.StatusOK,
			wantResponse:   `{"mutationInfos":[{"id":2,"error":"mutation id is too high - next expected mutation is: 1"}]}` + "\n",
			wantMutationID: 0,
			wantNumTodos:   0,
		},
		{
			label:          "missing-mutation-name",
			req:            `{"clientID":"c1","mutations":[{"id":1}]}`,
			userID:         1,
			wantCode:       http.StatusOK,
			wantResponse:   `{"mutationInfos":[{"id":1,"error":"mutation name is required"}]}` + "\n",
			wantMutationID: 1, // permanent error, so we update mutationID
			wantNumTodos:   0,
		},
		{
			label:          "unknown-mutation-name",
			req:            `{"clientID":"c1","mutations":[{"id":2,"name":"unknown-mutation"}]}`,
			userID:         1,
			wantCode:       http.StatusOK,
			wantResponse:   `{"mutationInfos":[{"id":2,"error":"unknown mutation name"}]}` + "\n",
			wantMutationID: 2,
			wantNumTodos:   0,
		},
		{
			label:          "mutator-permanent-error",
			req:            `{"clientID":"c1","mutations":[{"id":3,"name":"createTodo","args":{}}]}`,
			userID:         1,
			wantCode:       http.StatusOK,
			wantResponse:   `{"mutationInfos":[{"id":3,"error":"createTodo: id field is required"}]}` + "\n",
			wantMutationID: 3,
			wantNumTodos:   0,
		},
		{
			label:          "mutator-ok",
			req:            `{"clientID":"c1","mutations":[{"id":4,"name":"createTodo","args":{"id":1,"listID":1,"text":"text"}}]}`,
			userID:         1,
			wantCode:       http.StatusOK,
			wantResponse:   `{"mutationInfos":[]}` + "\n",
			wantMutationID: 4,
			wantNumTodos:   1,
		},
		{
			label:          "idempotency-3",
			req:            `{"clientID":"c1","mutations":[{"id":3}]}`,
			userID:         1,
			wantCode:       http.StatusOK,
			wantResponse:   `{"mutationInfos":[{"id":3,"error":"mutation has already been processed"}]}` + "\n",
			wantMutationID: 4,
			wantNumTodos:   1,
		},
		{
			label:          "multiple-mutators-ok",
			req:            `{"clientID":"c1","mutations":[{"id":5,"name":"createTodo","args":{"id":2,"listID":1,"text":"text"}},{"id":6,"name":"createTodo","args":{"id":3,"listID":1,"text":"text"}}]}`,
			userID:         1,
			wantCode:       http.StatusOK,
			wantResponse:   `{"mutationInfos":[]}` + "\n",
			wantMutationID: 6,
			wantNumTodos:   3,
		},
		{
			label: "multiple-mutators-skip",
			// second mutator is permanently invalid
			req:            `{"clientID":"c1","mutations":[{"id":7,"name":"createTodo","args":{"id":4,"listID":1,"text":"text"}},{"id":8},{"id":9,"name":"createTodo","args":{"id":5,"listID":1,"text":"text"}}]}`,
			userID:         1,
			wantCode:       http.StatusOK,
			wantResponse:   `{"mutationInfos":[{"id":8,"error":"mutation name is required"}]}` + "\n",
			wantMutationID: 9,
			wantNumTodos:   5,
		},
		{
			label: "multiple-mutators-stop",
			// second mutator skips id 11
			req:            `{"clientID":"c1","mutations":[{"id":10,"name":"createTodo","args":{"id":6,"listID":1,"text":"text"}},{"id":12},{"id":13,"name":"createTodo","args":{"id":7,"listID":1,"text":"text"}}]}`,
			userID:         1,
			wantCode:       http.StatusOK,
			wantResponse:   `{"mutationInfos":[{"id":12,"error":"mutation id is too high - next expected mutation is: 11"}]}` + "\n",
			wantMutationID: 10,
			wantNumTodos:   6,
		},
		// This test is basically a smoke test that expected registered mutators are present.
		// These mutators are tested in more detail in their directories.
		{
			label: "registered-mutators",
			req: `
{
	"clientID":"c1",
	"mutations":[
		{
			"id":11,
			"name":"createList",
			"args":{"id":2}
		},
		{
			"id":12,
			"name":"createTodo",
			"args":{"id":8,"listID":2,"text":"text"}
		},
		{
			"id":13,
			"name":"updateTodo",
			"args":{"id":8,"text":"text2"}
		},
		{
			"id":14,
			"name":"deleteTodo",
			"args":{"id":8}
		}
	]
}
`,
			userID:         1,
			wantCode:       http.StatusOK,
			wantResponse:   `{"mutationInfos":[]}` + "\n",
			wantMutationID: 14,
			wantNumTodos:   6,
		},
		// todo: transient errors from db/network
	}

	for _, t := range tc {
		msg := fmt.Sprintf("test case %s", t.label)
		w := httptest.NewRecorder()

		Handle(w, httptest.NewRequest("POST", "/replicache-batch", strings.NewReader(t.req)), db, userID)

		assert.Equal(t.wantCode, w.Result().StatusCode, msg)
		body := &bytes.Buffer{}
		_, err := io.Copy(body, w.Result().Body)
		assert.NoError(err, msg)
		assert.Equal(t.wantResponse, string(body.Bytes()), msg)

		gotMutationID, err := replicache.GetMutationID(db.Exec, "c1")
		assert.NoError(err, msg)
		assert.Equal(t.wantMutationID, gotMutationID, msg)

		ts, err := todo.GetByUser(db.Exec, 1)
		assert.NoError(err, msg)
		assert.Equal(t.wantNumTodos, len(ts), msg)
	}
}
