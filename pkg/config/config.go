package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Figma    FigmaConfig `yaml:"figma"`
	Schedule string      `yaml:"schedule"`
	Email    EmailConfig `yaml:"email"`
}

type FigmaConfig struct {
	Token    string   `yaml:"token"`
	FileKeys []string `yaml:"file_keys"`
}

type EmailConfig struct {
	SMTPHost     string   `yaml:"smtp_host"`
	SMTPPort     int      `yaml:"smtp_port"`
	SMTPUsername string   `yaml:"smtp_username"`
	SMTPPassword string   `yaml:"smtp_password"`
	From         string   `yaml:"from"`
	To           []string `yaml:"to"`
	Subject      string   `yaml:"subject"`
	Body         string   `yaml:"body"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}