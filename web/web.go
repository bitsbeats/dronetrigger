package web

import (
	"encoding/json"
	"fmt"
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
		Repo    string `json:"repo"`
		Branch  string `json:"branch"`
		Release bool   `json:"release_only"`
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
		WriteResponse(w, Response{
			StatusCode:  http.StatusInternalServerError,
			LogMsg:      fmt.Sprintf("unable to load request body as json: %s", err),
			ResponseMsg: "unable to parse request body",
		})
		return
	}
	if p.Repo == "" {
		WriteResponse(w, Response{
			StatusCode:  http.StatusInternalServerError,
			LogMsg:      "no repo specified",
			ResponseMsg: "no repo specified",
		})
		return
	}
	if _, ok := web.Config.BearerToken[p.Repo]; !ok {
		WriteResponse(w, Response{
			StatusCode:  http.StatusForbidden,
			LogMsg:      "invalid repository",
			ResponseMsg: "invalid repository",
		})
		return
	}
	if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %s", web.Config.BearerToken[p.Repo]) {
		WriteResponse(w, Response{
			StatusCode:  http.StatusForbidden,
			LogMsg:      "invalid bearer token",
			ResponseMsg: "invalid bearer token",
		})
		return
	}

	// handle request
	build := (*core.Build)(nil)
	if p.Release {
		build, err = web.Drone.RebuildLastTag(p.Repo)
	} else {
		build, err = web.Drone.RebuildLastBuild(p.Repo, p.Branch)
	}
	if err != nil || build == nil {
		WriteResponse(w, Response{
			StatusCode:  http.StatusInternalServerError,
			LogMsg:      fmt.Sprintf("unable to start last build for %s@%s: %s", p.Repo, p.Branch, err),
			ResponseMsg: "unable to restart build",
		})
		return
	}

	WriteResponse(w, Response{
		StatusCode:  http.StatusCreated,
		LogMsg:      fmt.Sprintf("started build %d %s@%s", build.Number, p.Repo, p.Branch),
		ResponseMsg: "ok",
	})
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

// Response is a helper to create uniform responses
type Response struct {
	StatusCode  int
	ResponseMsg string
	LogMsg      string
}

// WriteResponse writes a response to http.ResponseWriter
func WriteResponse(w http.ResponseWriter, r Response) {
	w.(*ResponseWriterWithStatus).SetMessage(r.LogMsg)
	w.WriteHeader(r.StatusCode)
	responseMsg := r.ResponseMsg
	errorMsg := ""
	if responseMsg == "" {
		responseMsg = "ok"
	}
	if r.StatusCode >= 400 {
		responseMsg = "error"
		errorMsg = r.ResponseMsg
	}
	jr := core.JsonResponse{
		Status: responseMsg,
		Err:    errorMsg,
	}
	_ = json.NewEncoder(w).Encode(jr)
}
