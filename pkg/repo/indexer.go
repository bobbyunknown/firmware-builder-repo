package repo

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Indexer struct {
	client *GitHubClient
}

func NewIndexer(owner, repo, branch string) *Indexer {
	return &Indexer{
		client: NewGitHubClient(owner, repo, branch),
	}
}

func (idx *Indexer) FetchKernelIndex() (*KernelsYAML, error) {
	contents, err := idx.client.ListContents("kernels")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch kernels: %w", err)
	}

	var kernels []KernelIndex
	for _, item := range contents {
		if item.Type == "dir" {
			vendor := "unknown"

			kernelContents, err := idx.client.ListContents(item.Path)
			if err == nil {
				for _, file := range kernelContents {
					if strings.HasPrefix(file.Name, "dtb-") && strings.HasSuffix(file.Name, ".tar.gz") {
						parts := strings.Split(file.Name, "-")
						if len(parts) >= 2 {
							vendor = parts[1]
						}
						break
					}
				}
			}

			kernels = append(kernels, KernelIndex{
				Version: item.Name,
				Vendor:  vendor,
				Path:    item.Path,
			})
		}
	}

	return &KernelsYAML{
		Metadata: IndexMetadata{
			Generated: time.Now().Format(time.RFC3339),
			Source:    fmt.Sprintf("%s/%s (branch: %s)", idx.client.Owner, idx.client.Repo, idx.client.Branch),
		},
		Kernels: kernels,
	}, nil
}

func (idx *Indexer) FetchRootfsIndex() (*RootfsYAML, error) {
	contents, err := idx.client.ListContents("rootfs")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rootfs: %w", err)
	}

	var rootfs []RootfsIndex
	for _, item := range contents {
		if item.Type == "file" && (item.Name != ".gitkeep" && item.Name != ".keep") {
			rootfsType := "base"
			if item.Size > 50*1024*1024 {
				rootfsType = "custom"
			}

			rootfs = append(rootfs, RootfsIndex{
				Name: item.Name,
				Type: rootfsType,
				Size: item.Size,
				Path: item.Path,
			})
		}
	}

	return &RootfsYAML{
		Metadata: IndexMetadata{
			Generated: time.Now().Format(time.RFC3339),
			Source:    fmt.Sprintf("%s/%s (branch: %s)", idx.client.Owner, idx.client.Repo, idx.client.Branch),
		},
		Rootfs: rootfs,
	}, nil
}

func (idx *Indexer) FetchDeviceIndex() (*DevicesYAML, error) {
	contents, err := idx.client.ListContents("devices")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch devices: %w", err)
	}

	var devices []DeviceIndex
	for _, item := range contents {
		if item.Type == "dir" {
			devices = append(devices, DeviceIndex{
				Name:   item.Name,
				Vendor: detectSOC(item.Name),
				Path:   item.Path,
			})
		}
	}

	return &DevicesYAML{
		Metadata: IndexMetadata{
			Generated: time.Now().Format(time.RFC3339),
			Source:    fmt.Sprintf("%s/%s (branch: %s)", idx.client.Owner, idx.client.Repo, idx.client.Branch),
		},
		Devices: devices,
	}, nil
}

func SaveIndex(path string, data interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)

	if err := encoder.Encode(data); err != nil {
		return err
	}

	return nil
}
