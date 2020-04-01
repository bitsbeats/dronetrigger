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
func (d *Drone) Builds(repo string) (builds []*core.Build, err error) {
	url := fmt.Sprintf("%s/api/repos/%s/builds", d.url, repo)
	builds = []*core.Build{}
	err = d.request("GET", url, nil, &builds)
	return
}

// Builds gets the last build for a specific branc
func (d *Drone) LastBuild(repo string, ref string) (b *core.Build, err error) {
	url := fmt.Sprintf("%s/api/repos/%s/builds/latest", d.url, repo)
	if ref != "" {
		url = fmt.Sprintf("%s?ref=%s", url, ref)
	}
	b = &core.Build{}
	err = d.request("GET", url, nil, b)
	return
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
func (d *Drone) RebuildLastBuild(repo string, ref string) (build *core.Build, err error) {
	lastBuild, err := d.LastBuild(repo, ref)
	if err != nil {
		return nil, err
	}
	build, err = d.Trigger(repo, lastBuild.Number)
	if err != nil {
		return
	}
	return
}

func (d *Drone) request(method string, url string, body io.Reader, result interface{}) (err error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return
	}
	if method == "POST" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", d.token))
	resp, err := d.client.Do(req)
	if err != nil {
		return
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if resp.StatusCode >= 300 {
		m, ok := result.(message)
		msg := resp.Status
		if ok && m.GetMessage() != "" {
			msg = fmt.Sprintf("%d %s", resp.StatusCode, m.GetMessage())
		}
		err = errors.New(msg)
		return
	}
	return
}
