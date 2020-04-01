package config

import (
	"testing"

	"github.com/bitsbeats/dronetrigger/core"
	check "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	check.TestingT(t)
}

type TestSuite struct{}

var _ = check.Suite(&TestSuite{})

func (s *TestSuite) TestConfig(c *check.C) {
	cfg, err := LoadConfig("test_files/with_web.yaml")
	c.Assert(err, check.DeepEquals, nil)
	c.Assert(cfg, check.DeepEquals, &core.Config{
		Url:   "https://drone.example.com",
		Token: "hi there",
		Web: &core.WebConfig{
			BearerToken: "bearer_token",
			Listen: ":8080",
		},
	})

	cfg, err = LoadConfig("test_files/with_web_and_port.yaml")
	c.Assert(err, check.DeepEquals, nil)
	c.Assert(cfg, check.DeepEquals, &core.Config{
		Url:   "https://drone.example.com",
		Token: "hi there",
		Web: &core.WebConfig{
			BearerToken: "bearer_token",
			Listen: ":1337",
		},
	})

	cfg, err = LoadConfig("test_files/without_web.yaml")
	c.Assert(err, check.DeepEquals, nil)
	c.Assert(cfg, check.DeepEquals, &core.Config{
		Url:   "https://drone.example.com",
		Token: "hi there",
		Web: nil,
	})

	cfg, err = LoadConfig("test_files/non-existent.yaml")
	c.Assert(err, check.ErrorMatches, "unable to open config: open test_files/non-existent.yaml: no such file or directory")
	c.Assert(cfg, check.Equals, (*core.Config)(nil))

	cfg, err = LoadConfig("test_files/invalid.yaml")
	c.Assert(err, check.ErrorMatches, "(?s)unable to parse config: yaml:.*")
	c.Assert(cfg, check.Equals, (*core.Config)(nil))
}
