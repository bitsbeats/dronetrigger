package drone

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/bitsbeats/dronetrigger/core"
)

type (
	// Drone is a API client for drone
	Drone struct {
		url    string
		token  string
		client *http.Client
	}

	message interface {
		GetMessage() string
	}

	BuildKind string
)

const (
	BUILD_PUSH BuildKind = "push"
	BUILD_TAG  BuildKind = "tag"
)

// New creates a new Drone API client
func New(url string, token string) *Drone {
	return &Drone{
		url:    url,
		token:  token,
		client: &http.Client{},
	}
}

// Builds lists all builds
func (d *Drone) Builds(repo string, page int) (builds []*core.Build, err error) {
	url := fmt.Sprintf("%s/api/repos/%s/builds?page=%d", d.url, repo, page)
	builds = []*core.Build{}
	err = d.request("GET", url, nil, &builds)
	if err != nil {
		return nil, err
	}
	return
}

// Builds gets the last build for a specific branc
func (d *Drone) LastBuild(repo string, branch string, kind BuildKind) (b *core.Build, err error) {
	if branch != "" {
		if kind == BUILD_TAG {
			return nil, fmt.Errorf("unable to build tag with branch filter")
		}
		url := fmt.Sprintf("%s/api/repos/%s/builds/latest?branch=%s", d.url, repo, branch)
		b = &core.Build{}
		err = d.request("GET", url, nil, b)
		if err != nil {
			return nil, err
		}
		return b, nil
	}

	// loop through pagination until the first matching build is found or error
	page := 0
	for b == nil {
		page += 1
		builds, err := d.Builds(repo, page)
		if err != nil {
			return nil, err
		}
		if len(builds) == 0 {
			break
		}
		for _, build := range builds {
			if BuildKind(build.Event) == kind {
				b = build
				break
			}
		}
	}
	if b == nil {
		return nil, fmt.Errorf("unable to find matching build")
	}
	return b, nil
}

// Trigger restarts a existing build by buildId
func (d *Drone) Trigger(repo string, buildId int64) (b *core.Build, err error) {
	url := fmt.Sprintf("%s/api/repos/%s/builds/%d?DRONETRIGGER=true", d.url, repo, buildId)
	b = &core.Build{}
	err = d.request("POST", url, nil, b)
	if err != nil {
		return nil, err
	}
	return
}

// RebuildLastBuild restarts the last build of a ref
func (d *Drone) RebuildLastBuild(repo string, branch string) (build *core.Build, err error) {
	lastBuild, err := d.LastBuild(repo, branch, BUILD_PUSH)
	if err != nil {
		return nil, err
	}
	build, err = d.Trigger(repo, lastBuild.Number)
	if err != nil {
		return nil, err
	}
	return
}

// RebuildLastTag restart the last tag build
func (d *Drone) RebuildLastTag(repo string) (build *core.Build, err error) {
	lastBuild, err := d.LastBuild(repo, "", BUILD_TAG)
	if err != nil {
		return nil, err
	}
	build, err = d.Trigger(repo, lastBuild.Number)
	if err != nil {
		return nil, err
	}
	return
}

func (d *Drone) request(method string, url string, body io.Reader, result interface{}) (err error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	if method == "POST" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", d.token))
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&result)
	if resp.StatusCode >= 400 {
		m, ok := result.(message)
		msg := resp.Status
		if ok && m.GetMessage() != "" {
			msg = fmt.Sprintf("%d %s", resp.StatusCode, m.GetMessage())
		}
		err = errors.New(msg)
		return err
	}
	return err
}
