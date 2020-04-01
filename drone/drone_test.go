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

func (s *TestSuite) TestFails(c *check.C) {
	mockCtrl := gomock.NewController(c)
	defer mockCtrl.Finish()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc("/api/repos/with/error/builds/latest", func(w http.ResponseWriter, r *http.Request){
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"message": "error description"}`)
	})
	server := httptest.NewServer(mux)
	d := New(server.URL, "")

	_, err := d.Builds("test/test")
	c.Assert(err, check.DeepEquals, fmt.Errorf("500 Internal Server Error"))

	_, err = d.LastBuild("test/test", "")
	c.Assert(err, check.DeepEquals, fmt.Errorf("500 Internal Server Error"))

	_, err = d.Trigger("test/test", 1337)
	c.Assert(err, check.DeepEquals, fmt.Errorf("500 Internal Server Error"))

	_, err = d.LastBuild("with/error", "")
	c.Assert(err, check.DeepEquals, fmt.Errorf("500 error description"))
}

func (s *TestSuite) TestHandler(c *check.C) {
	mockCtrl := gomock.NewController(c)
	defer mockCtrl.Finish()

	token := "q1QS0m6yFYRKm6TMPKeM8js8ZMbDLjPE"
	buildWasStarted := false
	mux := http.NewServeMux()
	mux.HandleFunc("/api/repos/bitsbeats/drone-test/builds", servJSON("test_files/builds.json", token))
	mux.HandleFunc("/api/repos/bitsbeats/drone-test/builds/latest", servJSON("test_files/latest.json", token))
	mux.HandleFunc("/api/repos/bitsbeats/drone-test/builds/58", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("DRONETRIGGER") != "true" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		servJSON("test_files/trigger.json", token)(w, r)
		buildWasStarted = true
	})
	server := httptest.NewServer(mux)

	d := New(server.URL, token)

	builds, err := d.Builds("bitsbeats/drone-test")
	buildsWant := []*core.Build{
		&core.Build{
			Message: "use alpine",
			Number:  58,
			Before:  "091a5a1f6afaa2148a447df71bad60f9f0518b56",
			After:   "a1e168b90d8ea1781ec73b84beedcad8e256e3fd",
			Source:  "master",
		},
		&core.Build{
			Message: "use alpine",
			Number:  57,
			Before:  "091a5a1f6afaa2148a447df71bad60f9f0518b56",
			After:   "a1e168b90d8ea1781ec73b84beedcad8e256e3fd",
			Source:  "master",
		},
		&core.Build{
			Message: "use alpine",
			Number:  56,
			Before:  "091a5a1f6afaa2148a447df71bad60f9f0518b56",
			After:   "a1e168b90d8ea1781ec73b84beedcad8e256e3fd",
			Source:  "master",
		},
	}
	c.Assert(err, check.Equals, nil)
	c.Assert(builds, check.DeepEquals, buildsWant)

	latest, err := d.LastBuild("bitsbeats/drone-test", "refs/heads/master")
	latestWant := buildsWant[0]
	c.Assert(err, check.Equals, nil)
	c.Assert(latest, check.DeepEquals, latestWant)

	build, err := d.RebuildLastBuild("bitsbeats/drone-test", "")
	buildWant := &core.Build{
			Message: "use alpine",
			Number:  59,
			Before:  "091a5a1f6afaa2148a447df71bad60f9f0518b56",
			After:   "a1e168b90d8ea1781ec73b84beedcad8e256e3fd",
			Source:  "master",
	}
	c.Assert(err, check.Equals, nil)
	c.Assert(build, check.DeepEquals, buildWant)
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
		io.Copy(w, fp)
	}
}
