package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// StarServer will start the HTTP server (blocking)
func StarServer() {
	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	log.Infof("Starting server at %s", addr)

	srv := &http.Server{
		Handler:      getRouter(),
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func getRouter() http.Handler {
	log.Infof("Creating HTTP router")
	return configureRouter(mux.NewRouter())
}

type handlerFunc interface {
	HandleFunc(path string, f func(http.ResponseWriter, *http.Request)) *mux.Route
	ServeHTTP(http.ResponseWriter, *http.Request)
}

func configureRouter(r handlerFunc) http.Handler {
	r.HandleFunc("/{api_version}/{rest:.*}", passthroughHandler)
	r.HandleFunc("/favicon.ico", notFoundHandler)
	r.HandleFunc("/{rest:.*}", passthroughHandler)
	r.HandleFunc("/", passthroughHandler)
	return r
}

// handles: /*
func passthroughHandler(w http.ResponseWriter, r *http.Request) {
	request := NewRequest(r, "passthrough", r.URL.String())
	request.log.Infof("Handling %s from %s", r.URL.String(), remoteIP(r.RemoteAddr))

	// publish specific go-metadataproxy headers
	request.setResponseHeaders(w)

	r.RequestURI = ""

	// ensure the chema and correct IP is set
	if r.URL.Scheme == "" {
		r.URL.Scheme = "http"
		r.URL.Host = "169.254.169.254"
		r.Host = "169.254.169.254"
	}

	// create HTTP client
	tp := newTransport()
	client := &http.Client{Transport: tp}

	// use the incoming http request to construct upstream request
	resp, err := client.Do(r)
	if err != nil {
		request.HandleError(err, 404, "not found", w)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

}

// handles: /favicon.ico
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}
