package httperr

import (
	"log"
	"net/http"
)

func ServerError(w http.ResponseWriter, logMsg string) {
	log.Printf("Error: %s", logMsg)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal Server Error"))
}

func ClientError(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(msg))
}
