package repo

type KernelIndex struct {
	Version string `yaml:"version"`
	Vendor  string `yaml:"vendor"`
	Path    string `yaml:"path"`
}

type RootfsIndex struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
	Size int64  `yaml:"size"`
	Path string `yaml:"path"`
}

type DeviceIndex struct {
	Name   string `yaml:"name"`
	Vendor string `yaml:"vendor"`
	Path   string `yaml:"path"`
}

type PatchIndex struct {
	Name string `yaml:"name"`
	Size int64  `yaml:"size"`
	Path string `yaml:"path"`
}

type IndexMetadata struct {
	Generated string `yaml:"generated"`
	Source    string `yaml:"source"`
}

type KernelsYAML struct {
	Metadata IndexMetadata `yaml:"metadata"`
	Kernels  []KernelIndex `yaml:"kernels"`
}

type RootfsYAML struct {
	Metadata IndexMetadata `yaml:"metadata"`
	Rootfs   []RootfsIndex `yaml:"rootfs"`
}

type DevicesYAML struct {
	Metadata IndexMetadata `yaml:"metadata"`
	Devices  []DeviceIndex `yaml:"devices"`
}

type PatchesYAML struct {
	Metadata IndexMetadata `yaml:"metadata"`
	Patches  []PatchIndex  `yaml:"patches"`
}
