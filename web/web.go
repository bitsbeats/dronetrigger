package web

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/bitsbeats/dronetrigger/core"
)

type (
	// Web is a WebAPI for dronetrigger
	Web struct {
		Config *core.WebConfig
		Drone  core.Drone
	}

	// Payload is the payload send to drone
	Payload struct {
		Repo string `json:"repo"`
		Ref  string `json:"ref"`
	}
)

// NewWeb creates a new Web
func NewWeb(c *core.WebConfig, d core.Drone) *Web {
	return &Web{
		Config: c,
		Drone:  d,
	}
}

// Handle handles an API request
func (web *Web) Handle(w http.ResponseWriter, r *http.Request) {
	// validate request
	p := Payload{}
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		w.(*ResponseWriterWithStatus).SetMessage(
			fmt.Sprintf("unable to load request body as json: %s", err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		Response(w, "error", fmt.Errorf("unable to parse request body"))
		return
	}
	if p.Repo == "" {
		w.(*ResponseWriterWithStatus).SetMessage("no repo specified")
		w.WriteHeader(http.StatusInternalServerError)
		Response(w, "error", fmt.Errorf("no repo specified"))
		return
	}
	if _, ok := web.Config.BearerToken[p.Repo]; !ok {
		w.(*ResponseWriterWithStatus).SetMessage("invalid repository")
		w.WriteHeader(http.StatusForbidden)
		Response(w, "error", fmt.Errorf("invalid repository"))
		return
	}
	if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %s", web.Config.BearerToken[p.Repo]) {
		w.(*ResponseWriterWithStatus).SetMessage("invalid bearer token")
		w.WriteHeader(http.StatusForbidden)
		Response(w, "error", fmt.Errorf("invalid bearer token"))
		return
	}

	// handle request
	build, err := web.Drone.RebuildLastBuild(p.Repo, p.Ref)
	if err != nil {
		w.(*ResponseWriterWithStatus).SetMessage(
			fmt.Sprintf("unable to start last build for %s@%s: %s", p.Repo, p.Ref, err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		Response(w, "error", fmt.Errorf("unable to restart build"))
		return
	}

	w.(*ResponseWriterWithStatus).SetMessage(
		fmt.Sprintf("started build %d %s@%s", build.Number, p.Repo, p.Ref),
	)
	w.WriteHeader(200)
	Response(w, "ok", nil)
}

// Middleware provides logging
func (web *Web) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		w = NewResponseWriterWithStatus(w)
		next.ServeHTTP(w, r)
		log.Printf(
			"%d %s %s %s %s - %s",
			w.(*ResponseWriterWithStatus).StatusCode,
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			time.Since(start).String(),
			w.(*ResponseWriterWithStatus).LogMessage,
		)
	})
}

// ResponseWriterWithStatus is a wrapper around ResponseWriter that tracks statusCode and LogMessage
type ResponseWriterWithStatus struct {
	http.ResponseWriter
	StatusCode int
	LogMessage string
}

// NewResponseWriterWithStatus creates a new ResponseWriter
func NewResponseWriterWithStatus(w http.ResponseWriter) *ResponseWriterWithStatus {
	return &ResponseWriterWithStatus{
		ResponseWriter: w,
		StatusCode:     200,
	}
}

// WriteHeader writes HTTP headers and stores the statusCode
func (r *ResponseWriterWithStatus) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// SetMessage sets the LogMessage
func (r *ResponseWriterWithStatus) SetMessage(message string) {
	r.LogMessage = message
}

// Response write a JsonResponse to
func Response(w io.Writer, status string, err error) {
	errText := ""
	if err != nil {
		errText = err.Error()
	}
	jr := core.JsonResponse{
		Status: status,
		Err:    errText,
	}
	_ = json.NewEncoder(w).Encode(jr)
}
