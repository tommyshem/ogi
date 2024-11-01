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

// TODO make multi repos to call offline to save going into repo folders
// Save writes the configuration to the .ogi.yml file in the current working
// directory. This should be called after setting the owner and repo fields.
func (c *Config) Save() {
	c.LastUpdated = time.Now()
	data, _ := yaml.Marshal(c)
	os.WriteFile("./.ogi.yml", data, 0755)
}

// SetFromArgs takes the first argument passed in and assumes it's a
// repository path in the format of "owner/repo". It splits the path
// and sets the owner and repo fields, then saves the configuration.
func (config *Config) SetFromArgs(args []string) {
	stringSplit := strings.Split(args[0], "/")
	if len(stringSplit) == 2 {
		config.Owner = stringSplit[0]
		config.Repo = stringSplit[1]
		config.Save()
	}
}

// LoadConfig loads the configuration from the .ogi.yml file in the current
// directory. If the file doesn't exist, or there's an error loading it, it
// returns a new Config object.
func LoadConfig() *Config {
	config := &Config{}
	data, err := os.ReadFile("./.ogi.yml")
	if err != nil {
		return config
	}
	yaml.Unmarshal(data, config)
	return config
}
