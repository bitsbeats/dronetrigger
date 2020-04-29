package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitsbeats/dronetrigger/core"
	"github.com/bitsbeats/dronetrigger/mock"
	"github.com/golang/mock/gomock"
	check "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	check.TestingT(t)
}

type TestSuite struct{}

var _ = check.Suite(&TestSuite{})

func (s *TestSuite) TestHandler(c *check.C) {
	mockCtrl := gomock.NewController(c)
	defer mockCtrl.Finish()

	tests := []struct {
		bearer string
		body   string
		repo   string
		branch string

		build    *core.Build
		droneErr error

		call bool
		resp *core.JsonResponse
	}{
		{
			bearer: "token",
			body:   `{"repo": "octocat/repo", "branch": "dev"}`,
			repo:   "octocat/repo", branch: "dev",

			build: &core.Build{Number: 1337}, droneErr: nil,
			call: true,
			resp: &core.JsonResponse{Status: "ok", Err: ""},
		},
		{
			bearer: "token",
			body:   `{"repo": "octocat/repo", "branch": ""}`,
			repo:   "octocat/repo", branch: "",

			build: &core.Build{Number: 1337}, droneErr: nil,

			call: true,
			resp: &core.JsonResponse{Status: "ok", Err: ""},
		},
		{
			bearer: "wrong",
			body:   `{"repo": "octocat/repo", "branch": "master"}`,
			repo:   "octocat/repo", branch: "master",

			build: &core.Build{Number: 1337}, droneErr: nil,

			call: false,
			resp: &core.JsonResponse{Status: "error", Err: "invalid bearer token"},
		},
		{
			bearer: "token",
			body:   `no json`,
			repo:   "octocat/repo", branch: "master",

			build: &core.Build{Number: 1337}, droneErr: nil,

			call: false,
			resp: &core.JsonResponse{Status: "error", Err: "unable to parse request body"},
		},
		{
			bearer: "token",
			body:   `{"branch": "master"}`,
			repo:   "octocat/repo", branch: "master",

			build: &core.Build{Number: 1337}, droneErr: nil,

			call: false,
			resp: &core.JsonResponse{Status: "error", Err: "no repo specified"},
		},
		{
			bearer: "token",
			body:   `{"repo": "octocat/repo", "branch": "master"}`,
			repo:   "octocat/repo", branch: "master",

			build: &core.Build{Number: 1337}, droneErr: fmt.Errorf("Fail"),

			call: true,
			resp: &core.JsonResponse{Status: "error", Err: "unable to restart build"},
		},
		{
			bearer: "token",
			body:   `{"repo": "octocat/repo2", "branch": "master"}`,
			repo:   "octocat/repo2", branch: "master",

			build: &core.Build{Number: 1337}, droneErr: nil,

			call: false,
			resp: &core.JsonResponse{Status: "error", Err: "invalid repository"},
		},
	}

	// test normal builds
	for _, test := range tests {
		d := mock.NewMockDrone(mockCtrl)
		if test.call {
			d.EXPECT().
				RebuildLastBuild(test.repo, test.branch).
				Return(test.build, test.droneErr)
		}

		web := NewWeb(&core.WebConfig{
			BearerToken: map[string]string{"octocat/repo": "token"},
			Listen:      ":1337",
		}, d)

		body := bytes.NewBufferString(test.body)
		r := httptest.NewRequest("POST", "/", body)
		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", test.bearer))
		w := NewResponseWriterWithStatus(httptest.NewRecorder())
		web.Handle(w, r)

		resp := &core.JsonResponse{}
		_ = json.NewDecoder(w.ResponseWriter.(*httptest.ResponseRecorder).Body).Decode(resp)
		c.Assert(*resp, check.Equals, *test.resp)

	}

	// test tag
	d := mock.NewMockDrone(mockCtrl)
	d.EXPECT().RebuildLastTag("octocat/repo3").Return(&core.Build{Number: 1337}, nil)
	web := NewWeb(&core.WebConfig{
		BearerToken: map[string]string{"octocat/repo3": "0ct0cat!"},
		Listen:      "1337",
	}, d)

	body := bytes.NewBufferString(`{"repo": "octocat/repo3", "release_only": true}`)
	r := httptest.NewRequest("POST", "/", body)
	r.Header.Set("Authorization", "Bearer 0ct0cat!")
	w := NewResponseWriterWithStatus(httptest.NewRecorder())
	web.Handle(w, r)

	resp := &core.JsonResponse{}
	_ = json.NewDecoder(w.ResponseWriter.(*httptest.ResponseRecorder).Body).Decode(resp)
	c.Assert(*resp, check.DeepEquals, core.JsonResponse{Status: "ok", Err: ""})

}

func (s *TestSuite) TestMiddleware(c *check.C) {
	mockCtrl := gomock.NewController(c)
	defer mockCtrl.Finish()

	d := mock.NewMockDrone(mockCtrl)
	web := NewWeb(&core.WebConfig{
		BearerToken: map[string]string{"octocat/repo": "token"},
		Listen:      ":1337",
	}, d)

	called := "middleware did not call handler"
	isResponseWriterWithStatus := false
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	middleware := web.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, isResponseWriterWithStatus = w.(*ResponseWriterWithStatus)
		called = "middleware called handler"
	}))
	middleware.ServeHTTP(w, r)

	c.Assert(called, check.Equals, "middleware called handler")
	if !c.Check(isResponseWriterWithStatus, check.Equals, true) {
		c.Fatalf("ResponseWriter was not changed to ResponseWriterWithStatus")
	}

}
