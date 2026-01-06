package repo

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

type RepoConfig struct {
	Version      string                     `yaml:"version"`
	Repositories map[string]RepositoryConfig `yaml:"repositories"`
}

type RepositoryConfig struct {
	Type        string            `yaml:"type"`
	URL         string            `yaml:"url"`
	Branch      string            `yaml:"branch"`
	Path        string            `yaml:"path"`
	CacheTTL    int               `yaml:"cache_ttl"`
	Description string            `yaml:"description"`
	Components  map[string]string `yaml:"components"`
}

func LoadConfig(configPath string) (*RepoConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg RepoConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func LoadDataRepo(configPath string) (*RepositoryConfig, error) {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	repoCfg, ok := cfg.Repositories["data"]
	if !ok {
		return nil, fmt.Errorf("missing repositories.data in %s", configPath)
	}

	return &repoCfg, nil
}

func ParseRepoURL(repoURL string) (string, string, error) {
	parsed, err := url.Parse(repoURL)
	if err != nil {
		return "", "", err
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid repo url: %s", repoURL)
	}
	return parts[0], parts[1], nil
}

func cleanComponentPath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	value = path.Clean("/" + value)
	return strings.TrimPrefix(value, "/")
}
