package config

import (
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Config struct {
	OpenAI struct {
		APIKey string `yaml:"api_key"`
		Model  string `yaml:"model"`
		Prompt string `yaml:"prompt"`
	} `yaml:"openai"`

	Gmail struct {
		ClientSecretPath string   `yaml:"client_secret_path"`
		TokenPath        string   `yaml:"token_path"`
		Labels           []string `yaml:"labels"`
	} `yaml:"gmail"`

	DB struct {
		Path string `yaml:"path"`
	} `yaml:"db"`
}

func LoadConfig(filePath string) (*Config, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read config file")
	}

	content = replaceEnvVars(content)

	var config Config
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal config file")
	}

	log.Println("Config:", config)
	return &config, nil
}

func replaceEnvVars(content []byte) []byte {
	envVarPattern := regexp.MustCompile(`\$\{(.+?)\}`)
	return envVarPattern.ReplaceAllFunc(content, func(s []byte) []byte {
		envVar := string(s[2 : len(s)-1])
		envValue := os.Getenv(envVar)
		return []byte(envValue)
	})
}
