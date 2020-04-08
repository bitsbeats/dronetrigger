package core

type (
	Config struct {
		Url   string     `yaml:"url"`
		Token string     `yaml:"token"`
		Web   *WebConfig `yaml:"web"`
	}

	WebConfig struct {
		BearerToken map[string]string `yaml:"bearer_token"`
		Listen      string            `yaml:"listen"`
	}
)
