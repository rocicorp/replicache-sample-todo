package mutator

import (
	"net/http"

	"roci.dev/replicache-sample-todo/serve/util/errs"
	"roci.dev/replicache-sample-todo/serve/util/httperr"
)

// Handle wraps a Replicache mutator implementation in a classic
// REST/HTTP interface. This is just so that code can be shared in
// this demo. There's no need to share code, or for the REST endpoint
// impls to have anything to do with Replicache.
func Handle(w http.ResponseWriter, mutator func() error) {
	err := mutator()
	switch err := err.(type) {
	case errs.BadRequestError:
		httperr.ClientError(w, err.Error())
	case errs.UnauthorizedError:
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
	default:
		httperr.ServerError(w, err.Error())
	}
}
