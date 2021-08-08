package main

import (
	"net/http"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

const (
	telemetryPrefix = "metadataproxy"
)

// Request ..
type Request struct {
	request       *http.Request
	vars          map[string]string
	id            string
	log           *logrus.Entry
	loggingLabels logrus.Fields
}

// NewRequest ..
func NewRequest(r *http.Request, name, path string) *Request {
	id := uuid.NewV4()

	// Create struct
	request := &Request{
		request:       r,
		vars:          mux.Vars(r),
		id:            id.String(),
		log:           logrus.WithField("request.id", id.String()),
		loggingLabels: logrus.Fields{},
	}

	return request
}

func (r *Request) setResponseHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Powered-By", "aws-proxy")
	w.Header().Set("X-Request-ID", r.id)
}

// HandleError ..
func (r *Request) HandleError(err error, code int, description string, w http.ResponseWriter) {
	r.log.Error(err)
	http.NotFound(w, nil)
}
