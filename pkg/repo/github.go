package repo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type GitHubClient struct {
	Owner  string
	Repo   string
	Branch string
	Token  string
}

type GitHubContent struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
	Size int64  `json:"size"`
}

func NewGitHubClient(owner, repo, branch string) *GitHubClient {
	return &GitHubClient{
		Owner:  owner,
		Repo:   repo,
		Branch: branch,
	}
}

func (c *GitHubClient) ListContents(path string) ([]GitHubContent, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
		c.Owner, c.Repo, path, c.Branch)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}

	var contents []GitHubContent
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return nil, err
	}

	return contents, nil
}

func detectSOC(name string) string {
	name = strings.ToLower(name)

	if strings.Contains(name, "s905") {
		return "amlogic"
	}
	if strings.Contains(name, "rk3") || strings.Contains(name, "rockchip") {
		return "rockchip"
	}
	if strings.Contains(name, "h6") || strings.Contains(name, "h616") ||
		strings.Contains(name, "h618") || strings.Contains(name, "h313") ||
		strings.Contains(name, "a64") || strings.Contains(name, "h5") ||
		strings.Contains(name, "aw64") {
		return "allwinner"
	}

	return "unknown"
}
