package cmd

import (
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Owner       string    `yaml:"owner"`
	Repo        string    `yaml:"repo"`
	LastUpdated time.Time `yaml:"last_updated"`
}

// Save writes the configuration to the .ghi.yml file in the current working
// directory. This should be called after setting the owner and repo fields.
func (c *Config) Save() {
	c.LastUpdated = time.Now()
	b, _ := yaml.Marshal(c)
	os.WriteFile("./.ghi.yml", b, 0755)
}

// SetFromArgs takes the first argument passed in and assumes it's a
// repository path in the format of "owner/repo". It splits the path
// and sets the owner and repo fields, then saves the configuration.
func (c *Config) SetFromArgs(args []string) {
	sp := strings.Split(args[0], "/")
	c.Owner = sp[0]
	c.Repo = sp[1]
	c.Save()
}

// LoadConfig loads the configuration from the .ghi.yml file in the current
// directory. If the file doesn't exist, or there's an error loading it, it
// returns a new Config object.
func LoadConfig() *Config {
	c := &Config{}
	b, err := os.ReadFile("./.ghi.yml")
	if err != nil {
		return c
	}

	yaml.Unmarshal(b, c)
	return c
}
