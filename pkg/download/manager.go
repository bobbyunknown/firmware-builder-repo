package download

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
)

type Manager struct {
	CacheDir string
	BaseURL  string
	Client   *http.Client
}

func NewManager(cacheDir, owner, repo, branch string) *Manager {
	baseURL := fmt.Sprintf("https://media.githubusercontent.com/media/%s/%s/%s", owner, repo, branch)

	return &Manager{
		CacheDir: cacheDir,
		BaseURL:  baseURL,
		Client:   &http.Client{},
	}
}

func (m *Manager) DownloadFile(remotePath, localPath string) error {
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	url := fmt.Sprintf("%s/%s", m.BaseURL, remotePath)

	resp, err := m.Client.Get(url)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		filepath.Base(localPath),
	)

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	fmt.Println()
	return nil
}

func (m *Manager) DownloadKernel(version string) error {
	kernelDir := filepath.Join(m.CacheDir, "kernels", version)

	if _, err := os.Stat(kernelDir); err == nil {
		fmt.Printf("Kernel %s already cached\n", version)
		return nil
	}

	fmt.Printf("Downloading kernel %s...\n", version)

	parts := strings.Split(m.BaseURL, "/")
	owner := parts[4]
	repo := parts[5]
	branch := parts[6]

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/kernels/%s?ref=%s",
		owner, repo, version, branch)

	resp, err := m.Client.Get(apiURL)
	if err != nil {
		return fmt.Errorf("failed to list kernel files: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("kernel not found: HTTP %d", resp.StatusCode)
	}

	var files []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	for _, file := range files {
		if file.Type == "file" {
			remotePath := fmt.Sprintf("kernels/%s/%s", version, file.Name)
			localPath := filepath.Join(kernelDir, file.Name)

			if err := m.DownloadFile(remotePath, localPath); err != nil {
				fmt.Printf("Warning: failed to download %s: %v\n", file.Name, err)
			}
		}
	}

	return nil
}

func (m *Manager) DownloadRootfs(name string) error {
	rootfsDir := filepath.Join(m.CacheDir, "rootfs")
	localPath := filepath.Join(rootfsDir, name)

	if _, err := os.Stat(localPath); err == nil {
		fmt.Printf("Rootfs %s already cached\n", name)
		return nil
	}

	fmt.Printf("Downloading rootfs %s...\n", name)

	remotePath := fmt.Sprintf("rootfs/%s", name)

	return m.DownloadFile(remotePath, localPath)
}

func (m *Manager) GetKernelPath(version string) string {
	return filepath.Join(m.CacheDir, "kernels", version)
}

func (m *Manager) GetRootfsPath(name string) string {
	return filepath.Join(m.CacheDir, "rootfs", name)
}

func (m *Manager) IsCached(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
