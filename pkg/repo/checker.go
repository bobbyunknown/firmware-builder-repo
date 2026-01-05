package repo

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type DataChecker struct {
	Owner  string
	Repo   string
	Branch string
	Client *http.Client
}

func NewDataChecker(owner, repo, branch string) *DataChecker {
	return &DataChecker{
		Owner:  owner,
		Repo:   repo,
		Branch: branch,
		Client: &http.Client{},
	}
}

func (dc *DataChecker) CheckPath(path string) (bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
		dc.Owner, dc.Repo, path, dc.Branch)

	resp, err := dc.Client.Get(url)
	if err != nil {
		return false, fmt.Errorf("failed to check path: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return false, nil
	}

	if resp.StatusCode != 200 {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return true, nil
}

func (dc *DataChecker) ListDirectory(path string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
		dc.Owner, dc.Repo, path, dc.Branch)

	resp, err := dc.Client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to list directory: HTTP %d", resp.StatusCode)
	}

	var items []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var names []string
	for _, item := range items {
		names = append(names, item.Name)
	}

	return names, nil
}

func (dc *DataChecker) CheckLoader(vendor string) ([]string, error) {
	path := fmt.Sprintf("loader/%s", vendor)
	return dc.ListDirectory(path)
}

func (dc *DataChecker) CheckFirmware() (bool, error) {
	return dc.CheckPath("firmware")
}

func (dc *DataChecker) CheckDevices() ([]string, error) {
	return dc.ListDirectory("devices")
}

func (dc *DataChecker) VerifyDataStructure() error {
	fmt.Println("Verifying data repository structure...")

	components := []string{"kernels", "rootfs", "firmware", "devices", "loader"}

	for _, comp := range components {
		exists, err := dc.CheckPath(comp)
		if err != nil {
			return fmt.Errorf("failed to check %s: %w", comp, err)
		}
		if !exists {
			return fmt.Errorf("missing component: %s", comp)
		}
		fmt.Printf("   ✓ %s/ exists\n", comp)
	}

	vendors := []string{"amlogic", "allwinner", "rockchip"}
	for _, vendor := range vendors {
		loaders, err := dc.CheckLoader(vendor)
		if err != nil {
			fmt.Printf("   ⚠ loader/%s/ check failed: %v\n", vendor, err)
			continue
		}
		fmt.Printf("   ✓ loader/%s/ (%d files)\n", vendor, len(loaders))
	}

	return nil
}
