package github_exporter //nolint:golint

import (
  "os"
  "strings"

	"github.com/go-kit/kit/log"
  "github.com/infinityworks/github-exporter/exporter"
  gh_config "github.com/infinityworks/github-exporter/config"
  "github.com/grafana/agent/pkg/integrations"
  "github.com/grafana/agent/pkg/integrations/config"
)

var DefaultConfig Config = Config{
  ApiUrl: "https://api.github.com",
}

type Config struct {
  Common config.Common `yaml:"inline"`

  ApiUrl string `yaml:"api_url,omitempty"`

  Repositories []string `yaml:"repositories,omitempty"`

  Organizations []string `yaml:"organizations,omitempty"`

  Users []string `yaml:"users,omitempty"`

  ApiToken string `yaml:"api_token,omitempty"`

  ApiTokenFile string `yaml:"api_token_file,omitempty"`
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
  *c = DefaultConfig

  type plain Config
  return unmarshal((*plain)(c))
}

func (c *Config) Name() string {
  return "github_exporter"
}

func (c *Config) CommonConfig() config.Common {
  return c.Common
}

func (c *Config) NewIntegration(logger log.Logger) (integrations.Integration, error) {
  return New(logger, c)
}

func init() {
  integrations.RegisterIntegration(&Config{})
}

func New(logger log.Logger, c *Config) (integrations.Integration, error) {
  // It's not very pretty, but this exporter is configured entirely by environment
  // variables, and uses some private helper methods in it's config package to
  // assemble other key pieces of the config. Thus, we can't (easily) access the
  // config directly, and must assign environment variables.
  //
  // In an effort to avoid conflicts with other integrations, the environment vars
  // are unset immediately after being consumed.

  os.Setenv("API_URL", c.ApiUrl)
  os.Setenv("REPOS", strings.Join(c.Repositories, " "))
  os.Setenv("ORGS", strings.Join(c.Organizations, " "))
  os.Setenv("USERS", strings.Join(c.Users, " "))
  if c.ApiToken != "" {
    os.Setenv("GITHUB_TOKEN", c.ApiToken)
  }

  if c.ApiTokenFile != "" {
    os.Setenv("GITHUB_TOKEN_FILE", c.ApiTokenFile)
  }

  conf := gh_config.Init()

  os.Unsetenv("API_URL")
  os.Unsetenv("REPOS")
  os.Unsetenv("ORGS")
  os.Unsetenv("USERS")
  os.Unsetenv("GITHUB_TOKEN")
  os.Unsetenv("GITHUB_TOKEN_FILE")

  gh_exporter := exporter.Exporter{
    APIMetrics: exporter.AddMetrics(),
    Config: conf,
  }

  return integrations.NewCollectorIntegration(
    c.Name(),
    integrations.WithCollectors(&gh_exporter),
  ), nil
}
