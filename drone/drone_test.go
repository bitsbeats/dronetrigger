package drone

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/bitsbeats/dronetrigger/core"
	"github.com/golang/mock/gomock"
	check "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	check.TestingT(t)
}

type TestSuite struct{}

var _ = check.Suite(&TestSuite{})

func (s *TestSuite) TestApiFails(c *check.C) {
	// invalid calls
	mockCtrl := gomock.NewController(c)
	defer mockCtrl.Finish()

	server := httptest.NewServer(nil)
	d := New(server.URL, "")

	_, err := d.LastBuild("test/test", "master", BUILD_TAG)
	c.Assert(err, check.DeepEquals, fmt.Errorf("unable to build tag with branch filter"))
}

func (s *TestSuite) TestFails(c *check.C) {
	mockCtrl := gomock.NewController(c)
	defer mockCtrl.Finish()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc("/api/repos/with/error/builds/42", func(w http.ResponseWriter, r *http.Request) {
		c.Logf("url: %s", r.RequestURI)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"message": "Error description"}`)
	})
	mux.HandleFunc("/api/repos/not/found/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	server := httptest.NewServer(mux)
	d := New(server.URL, "")

	// 5xx
	_, err := d.Builds("test/test", 1)
	c.Assert(err, check.DeepEquals, fmt.Errorf("500 Internal Server Error"))

	_, err = d.LastBuild("test/test", "", BUILD_PUSH)
	c.Assert(err, check.DeepEquals, fmt.Errorf("500 Internal Server Error"))

	_, err = d.Trigger("test/test", 1337)
	c.Assert(err, check.DeepEquals, fmt.Errorf("500 Internal Server Error"))

	_, err = d.Trigger("with/error", 42)
	c.Assert(err, check.DeepEquals, fmt.Errorf("500 Error description"))

	// 404s
	_, err = d.LastBuild("not/found", "", BUILD_PUSH)
	c.Assert(err, check.DeepEquals, fmt.Errorf("404 Not Found"))

	_, err = d.LastBuild("not/found", "master", BUILD_PUSH)
	c.Assert(err, check.DeepEquals, fmt.Errorf("404 Not Found"))

	_, err = d.Builds("not/found", 1)
	c.Assert(err, check.DeepEquals, fmt.Errorf("404 Not Found"))

	_, err = d.Trigger("not/found", 23)
	c.Assert(err, check.DeepEquals, fmt.Errorf("404 Not Found"))

}

func (s *TestSuite) TestHandler(c *check.C) {
	mockCtrl := gomock.NewController(c)
	defer mockCtrl.Finish()

	token := "q1QS0m6yFYRKm6TMPKeM8js8ZMbDLjPE"
	buildWasStarted := false
	mux := http.NewServeMux()
	mux.HandleFunc("/api/repos/bitsbeats/drone-test/builds", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.FormValue("page") == "1" {
			servJSON("test_files/builds.json", token)(w, r)
		} else {
			fmt.Fprintf(w, "[]")
		}
	})
	mux.HandleFunc("/api/repos/bitsbeats/drone-test/builds/latest", servJSON("test_files/latest.json", token))
	mux.HandleFunc("/api/repos/bitsbeats/drone-test/builds/58", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Query().Get("DRONETRIGGER") != "true" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		servJSON("test_files/trigger.json", token)(w, r)
		buildWasStarted = true
	})
	server := httptest.NewServer(mux)

	d := New(server.URL, token)

	builds, err := d.Builds("bitsbeats/drone-test", 1)
	buildsWant := []*core.Build{
		&core.Build{
			Message: "use alpine",
			Number:  58,
			Before:  "091a5a1f6afaa2148a447df71bad60f9f0518b56",
			After:   "a1e168b90d8ea1781ec73b84beedcad8e256e3fd",
			Source:  "master",
			Event:   "push",
		},
		&core.Build{
			Message: "use alpine",
			Number:  57,
			Before:  "091a5a1f6afaa2148a447df71bad60f9f0518b56",
			After:   "a1e168b90d8ea1781ec73b84beedcad8e256e3fd",
			Source:  "master",
			Event:   "push",
		},
		&core.Build{
			Message: "use alpine",
			Number:  56,
			Before:  "091a5a1f6afaa2148a447df71bad60f9f0518b56",
			After:   "a1e168b90d8ea1781ec73b84beedcad8e256e3fd",
			Source:  "master",
			Event:   "push",
		},
	}
	c.Assert(err, check.Equals, nil)
	c.Assert(builds, check.DeepEquals, buildsWant)

	// check list builds
	latest, err := d.LastBuild("bitsbeats/drone-test", "master", BUILD_PUSH)
	latestWant := buildsWant[0]
	c.Assert(err, check.Equals, nil)
	c.Assert(latest, check.DeepEquals, latestWant)

	// check list builds for non-existing build tags
	latest, err = d.LastBuild("bitsbeats/drone-test", "", BUILD_TAG)
	c.Assert(err, check.DeepEquals, fmt.Errorf("unable to find matching build"))
	c.Assert(latest, check.Equals, (*core.Build)(nil))

	// check rebuild last build
	c.Assert(buildWasStarted, check.Equals, false) // no one should have restared by now
	build, err := d.RebuildLastBuild("bitsbeats/drone-test", "")
	buildWant := &core.Build{
		Message: "use alpine",
		Number:  59,
		Before:  "091a5a1f6afaa2148a447df71bad60f9f0518b56",
		After:   "a1e168b90d8ea1781ec73b84beedcad8e256e3fd",
		Source:  "master",
		Event:   "push",
	}
	c.Assert(err, check.Equals, nil)
	c.Assert(build, check.DeepEquals, buildWant)
	c.Assert(buildWasStarted, check.Equals, true)
}

func (s *TestSuite) TestTagHandler(c *check.C) {
	mockCtrl := gomock.NewController(c)
	defer mockCtrl.Finish()

	buildWasStarted := false
	token := "q1QS0m6yFYRKm6TMPKeM8js8ZMbDLjPE"
	mux := http.NewServeMux()
	mux.HandleFunc("/api/repos/bitsbeats/drone-test/builds", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.FormValue("page") == "1" {
			servJSON("test_files/builds_with_tag.json", token)(w, r)
		} else {
			fmt.Fprintf(w, "[]")
		}
	})
	mux.HandleFunc("/api/repos/bitsbeats/drone-test/builds/62", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Query().Get("DRONETRIGGER") != "true" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		servJSON("test_files/trigger_with_tag.json", token)(w, r)
		buildWasStarted = true
	})
	server := httptest.NewServer(mux)
	d := New(server.URL, token)

	buildWant := &core.Build{
		Message: "use alpine",
		Number:  62,
		Before:  "0000000000000000000000000000000000000000",
		After:   "a1e168b90d8ea1781ec73b84beedcad8e256e3fd",
		Source:  "master",
		Event:   "tag",
	}

	// just find last build
	latest, err := d.LastBuild("bitsbeats/drone-test", "", BUILD_TAG)
	c.Assert(err, check.Equals, nil)
	c.Assert(latest, check.DeepEquals, buildWant)

	// restart last build
	buildWant.Number = 64
	c.Assert(buildWasStarted, check.Equals, false) // no one should have started a build
	latest, err = d.RebuildLastTag("bitsbeats/drone-test")
	c.Assert(err, check.Equals, nil)
	c.Assert(latest, check.DeepEquals, buildWant)
	c.Assert(buildWasStarted, check.Equals, true)
}

func servJSON(path, token string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %s", token) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fp, _ := os.Open(path)
		_, _ = io.Copy(w, fp)
	}
}
