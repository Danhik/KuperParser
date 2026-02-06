package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Kuper struct {
		BaseURL string `yaml:"base_url"`
		StoreID int    `yaml:"store_id"`
	} `yaml:"kuper"`

	Departments struct {
		Names []string `yaml:"names"`
	} `yaml:"departments"`

	Pagination struct {
		PerPage     int `yaml:"per_page"`
		OffersLimit int `yaml:"offers_limit"`
	} `yaml:"pagination"`

	HTTP struct {
		TimeoutSeconds int `yaml:"timeout_seconds"`
		Retries        int `yaml:"retries"`
	} `yaml:"http"`

	Proxy struct {
		Mode        string   `yaml:"mode"`
		List        []string `yaml:"list"`
		RotationURL string   `yaml:"rotation_url"`
	} `yaml:"proxy"`

	Concurrency struct {
		Workers int `yaml:"workers"`
	} `yaml:"concurrency"`

	Output struct {
		Directory string `yaml:"directory"`
		Format    string `yaml:"format"`
	} `yaml:"output"`
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
