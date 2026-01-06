package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version      string                `yaml:"version"`
	Repositories map[string]Repository `yaml:"repositories"`
}

type Repository struct {
	Type        string            `yaml:"type"`
	URL         string            `yaml:"url"`
	Branch      string            `yaml:"branch"`
	Path        string            `yaml:"path"`
	CacheTTL    int               `yaml:"cache_ttl"`
	Description string            `yaml:"description"`
	Components  map[string]string `yaml:"components"`
}

type DeviceIndex struct {
	Metadata struct {
		Generated string `yaml:"generated"`
		Source    string `yaml:"source"`
	} `yaml:"metadata"`
	Devices []Device `yaml:"devices"`
}

type Device struct {
	Name   string `yaml:"name"`
	Vendor string `yaml:"vendor"`
	Path   string `yaml:"path"`
}

type KernelIndex struct {
	Metadata struct {
		Generated string `yaml:"generated"`
		Source    string `yaml:"source"`
	} `yaml:"metadata"`
	Kernels []Kernel `yaml:"kernels"`
}

type Kernel struct {
	Version string `yaml:"version"`
	Vendor  string `yaml:"vendor"`
	Path    string `yaml:"path"`
}

type RootfsIndex struct {
	Metadata struct {
		Generated string `yaml:"generated"`
		Source    string `yaml:"source"`
	} `yaml:"metadata"`
	Rootfs []Rootfs `yaml:"rootfs"`
}

type Rootfs struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

func (r *Repository) CacheDir() string {
	return filepath.Join(".cache", "data")
}

func Load() (*Config, error) {
	configPath := "configs/config.yaml"

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

func LoadDevices() (*DeviceIndex, error) {
	data, err := os.ReadFile("configs/devices.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read devices.yaml: %w", err)
	}

	var index DeviceIndex
	if err := yaml.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse devices.yaml: %w", err)
	}

	return &index, nil
}

func LoadKernels() (*KernelIndex, error) {
	data, err := os.ReadFile("configs/kernels.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read kernels.yaml: %w", err)
	}

	var index KernelIndex
	if err := yaml.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse kernels.yaml: %w", err)
	}

	return &index, nil
}

func LoadRootfs() (*RootfsIndex, error) {
	data, err := os.ReadFile("configs/rootfs.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read rootfs.yaml: %w", err)
	}

	var index RootfsIndex
	if err := yaml.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse rootfs.yaml: %w", err)
	}

	return &index, nil
}

func GetDeviceVendor(deviceName string) (string, error) {
	devices, err := LoadDevices()
	if err != nil {
		return "", err
	}

	deviceName = strings.ToLower(deviceName)
	for _, device := range devices.Devices {
		if strings.ToLower(device.Name) == deviceName {
			return device.Vendor, nil
		}
		if strings.Contains(deviceName, strings.ToLower(device.Name)) {
			return device.Vendor, nil
		}
	}

	return "", fmt.Errorf("device not found: %s", deviceName)
}
