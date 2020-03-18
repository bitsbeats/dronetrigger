package drone

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type (
	Drone struct {
		url    string
		token  string
		client *http.Client
	}
	message interface {
		GetMessage() string
	}
)

func New(url string, token string) *Drone {
	return &Drone{
		url:    url,
		token:  token,
		client: &http.Client{},
	}
}

func (d *Drone) Builds(repo string) (builds []Build, err error) {
	url := fmt.Sprintf("%s/api/repos/%s/builds", d.url, repo)
	err = d.request("GET", url, nil, &builds)
	return
}

func (d *Drone) LastBuild(repo string, ref string) (b Build, err error) {
	url := fmt.Sprintf("%s/api/repos/%s/builds/latest", d.url, repo)
	if ref != "" {
		url = fmt.Sprintf("%s?ref=%s", url, ref)
	}
	err = d.request("GET", url, nil, &b)
	return
}

func (d *Drone) Trigger(repo string, branch string, sha string) (b Build, err error) {
	url := fmt.Sprintf("%s/api/repos/%s/builds?DRONETRIGGER=true", d.url, repo)
	body := strings.NewReader(fmt.Sprintf("commit=%s&branch=%s", sha, branch))
	err = d.request("POST", url, body, &b)
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
		m, ok := (result).(message)
		msg := fmt.Sprintf("%d %s", resp.StatusCode, m.GetMessage())
		if !ok {
			msg = resp.Status
		}
		err = errors.New(msg)
		return
	}
	return
}
