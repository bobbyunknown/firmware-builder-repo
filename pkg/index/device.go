package index

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Device struct {
	Name        string `yaml:"name"`
	Vendor      string `yaml:"vendor"`
	Path        string `yaml:"path"`
	Description string `yaml:"description,omitempty"`
}

type DeviceRegistry struct {
	Devices []Device `yaml:"devices"`
}

func LoadDevices(path string) (*DeviceRegistry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var registry DeviceRegistry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return nil, err
	}

	return &registry, nil
}

func (r *DeviceRegistry) FindByName(name string) *Device {
	for _, device := range r.Devices {
		if device.Name == name {
			return &device
		}
	}
	return nil
}

func (r *DeviceRegistry) FindByVendor(vendor string) []Device {
	var devices []Device
	for _, device := range r.Devices {
		if device.Vendor == vendor {
			devices = append(devices, device)
		}
	}
	return devices
}

func (r *DeviceRegistry) ListAll() []Device {
	return r.Devices
}
