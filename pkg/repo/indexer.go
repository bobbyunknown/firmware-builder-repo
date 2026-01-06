package repo

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Indexer struct {
	client     *GitHubClient
	components map[string]string
}

func NewIndexer(owner, repo, branch string, components map[string]string) *Indexer {
	return &Indexer{
		client:     NewGitHubClient(owner, repo, branch),
		components: components,
	}
}

func (idx *Indexer) FetchKernelIndex() (*KernelsYAML, error) {
	contents, err := idx.client.ListContents(idx.componentPath("kernels"))
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
	contents, err := idx.client.ListContents(idx.componentPath("rootfs"))
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
	contents, err := idx.client.ListContents(idx.componentPath("devices"))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch devices: %w", err)
	}

	var devices []DeviceIndex
	for _, item := range contents {
		if item.Type == "dir" {
			vendor := "unknown"

			parts := strings.Split(item.Name, "-")
			if len(parts) > 0 {
				soc := parts[0]
				if strings.Contains(soc, "s905") || strings.Contains(soc, "s912") || strings.Contains(soc, "s922") || strings.Contains(soc, "a311") {
					vendor = "amlogic"
				} else if strings.Contains(soc, "rk3") {
					vendor = "rockchip"
				} else if strings.Contains(soc, "h") || strings.Contains(soc, "a64") {
					vendor = "allwinner"
				}
			}

			devices = append(devices, DeviceIndex{
				Name:   item.Name,
				Vendor: vendor,
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

func (idx *Indexer) FetchPatchIndex() (*PatchesYAML, error) {
	contents, err := idx.client.ListContents(idx.componentPath("patch"))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch patch: %w", err)
	}

	var patches []PatchIndex
	for _, item := range contents {
		if item.Type == "file" && (item.Name != ".gitkeep" && item.Name != ".keep") {
			patches = append(patches, PatchIndex{
				Name: item.Name,
				Size: item.Size,
				Path: item.Path,
			})
		}
	}

	return &PatchesYAML{
		Metadata: IndexMetadata{
			Generated: time.Now().Format(time.RFC3339),
			Source:    fmt.Sprintf("%s/%s (branch: %s)", idx.client.Owner, idx.client.Repo, idx.client.Branch),
		},
		Patches: patches,
	}, nil
}

func (idx *Indexer) componentPath(name string) string {
	if idx.components == nil {
		return name
	}
	if value, ok := idx.components[name]; ok {
		return cleanComponentPath(value)
	}
	return name
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
