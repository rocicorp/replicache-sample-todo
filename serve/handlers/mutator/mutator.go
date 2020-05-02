package mutator

import (
	"net/http"

	"roci.dev/replicache-sample-todo/serve/util/errs"
	"roci.dev/replicache-sample-todo/serve/util/httperr"
)

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
