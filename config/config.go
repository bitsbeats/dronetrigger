package config

import (
	"fmt"
	"io/ioutil"

	"github.com/bitsbeats/dronetrigger/core"
	"gopkg.in/yaml.v2"
)

func LoadConfig(path string) (c *core.Config, err error) {
	configData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open config: %w", err)
	}

	c = &core.Config{}
	err = yaml.Unmarshal(configData, c)
	if err != nil {
		return nil, fmt.Errorf("unable to parse config: %w", err)
	}

	if c.Web != nil && (c.Web.Listen == "") {
		c.Web.Listen = ":8080"
	}
	return
}
