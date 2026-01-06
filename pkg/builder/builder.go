package builder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bobbyunknown/Oh-my-builder/pkg/config"
	"github.com/bobbyunknown/Oh-my-builder/pkg/download"
	"github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/disk"
	"github.com/diskfs/go-diskfs/filesystem"
	"github.com/diskfs/go-diskfs/partition/mbr"
)

type BuildConfig struct {
	Device string
	Kernel string
	Rootfs string
	Size   int
	Output string
}

type Builder struct {
	Config   BuildConfig
	CacheDir string
	TempDir  string
	WorkDir  string
}

func NewBuilder(config BuildConfig, cacheDir string) (*Builder, error) {
	tempDir := filepath.Join("tmp", "build")
	workDir := filepath.Join(tempDir, "work")

	if _, err := os.Stat(tempDir); err == nil {
		fmt.Println("Cleaning up previous build directory...")
		if err := os.RemoveAll(tempDir); err != nil {
			return nil, fmt.Errorf("cleanup previous build: %w", err)
		}
	}

	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("create work directory: %w", err)
	}

	return &Builder{
		Config:   config,
		CacheDir: cacheDir,
		TempDir:  tempDir,
		WorkDir:  workDir,
	}, nil
}

func (b *Builder) Build() error {
	fmt.Println("🔨 Building firmware image...")
	fmt.Printf("   Device: %s\n", b.Config.Device)
	fmt.Printf("   Kernel: %s\n", b.Config.Kernel)
	fmt.Printf("   Rootfs: %s\n", b.Config.Rootfs)
	fmt.Printf("   Size: %d MB\n", b.Config.Size)
	fmt.Printf("   Output: %s\n\n", b.Config.Output)

	if err := b.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := b.CreateImage(); err != nil {
		return fmt.Errorf("create image failed: %w", err)
	}

	if err := b.InstallKernel(); err != nil {
		return fmt.Errorf("install kernel failed: %w", err)
	}

	if err := b.InstallRootfs(); err != nil {
		return fmt.Errorf("install rootfs failed: %w", err)
	}

	if err := b.WriteBootloader(); err != nil {
		return fmt.Errorf("write bootloader failed: %w", err)
	}

	fmt.Println("\n✅ Firmware image built successfully!")
	fmt.Printf("   Output: %s\n", b.Config.Output)

	return nil
}

func (b *Builder) Validate() error {
	fmt.Println("Checking resources...")

	dm, err := download.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create download manager: %w", err)
	}

	kernelPath := dm.GetKernelPath(b.Config.Kernel)
	if _, err := os.Stat(kernelPath); os.IsNotExist(err) {
		fmt.Printf("   Kernel %s not found locally. Auto-downloading...\n", b.Config.Kernel)
		if err := dm.DownloadKernel(b.Config.Kernel); err != nil {
			return fmt.Errorf("failed to auto-download kernel: %w", err)
		}
	} else {
		fmt.Printf("   Kernel %s available\n", b.Config.Kernel)
	}

	rootfsPath := dm.GetRootfsPath(b.Config.Rootfs)
	if _, err := os.Stat(rootfsPath); os.IsNotExist(err) {
		fmt.Printf("   Rootfs %s not found locally. Auto-downloading...\n", b.Config.Rootfs)
		if err := dm.DownloadRootfs(b.Config.Rootfs); err != nil {
			return fmt.Errorf("failed to auto-download rootfs: %w", err)
		}
	} else {
		fmt.Printf("   Rootfs %s available\n", b.Config.Rootfs)
	}

	return nil
}

func (b *Builder) CreateImage() error {
	fmt.Println("💾 Creating disk image...")

	imagePath := b.Config.Output
	// Total image = 16MB (bootloader space) + 256MB (boot partition) + rootfs size
	// This matches ulo script: fallocate -l $((16 + 256 + rootsize))M
	imageSize := int64(16+256+b.Config.Size) * 1024 * 1024

	if err := os.MkdirAll(filepath.Dir(imagePath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Remove existing image if it exists, ignore error if it doesn't exist
	os.Remove(imagePath)

	mydisk, err := diskfs.Create(imagePath, imageSize, diskfs.SectorSizeDefault)
	if err != nil {
		return fmt.Errorf("failed to create disk image: %w", err)
	}

	table := &mbr.Table{
		Partitions: []*mbr.Partition{
			{
				Bootable: true,
				Type:     mbr.Fat32LBA,
				Start:    2048,
				Size:     uint32(256 * 1024 * 1024 / 512),
			},
			{
				Bootable: false,
				Type:     mbr.Linux,
				Start:    2048 + uint32(256*1024*1024/512),
				Size:     uint32((imageSize - 257*1024*1024) / 512),
			},
		},
	}

	if err := mydisk.Partition(table); err != nil {
		return fmt.Errorf("failed to write partition table: %w", err)
	}

	fmt.Println("   Formatting boot partition (FAT32)...")
	_, err = mydisk.CreateFilesystem(disk.FilesystemSpec{
		Partition:   1,
		FSType:      filesystem.TypeFat32,
		VolumeLabel: "BOOT",
	})
	if err != nil {
		return fmt.Errorf("failed to create boot filesystem: %w", err)
	}

	fmt.Println("   ✓ Disk image created with partitions")
	fmt.Println("   Note: Rootfs partition will be formatted during installation")
	return nil
}

func (b *Builder) WriteBootloader() error {
	fmt.Println("🚀 Writing bootloader...")

	vendor, err := config.GetDeviceVendor(b.Config.Device)
	if err != nil {
		return fmt.Errorf("failed to detect vendor: %w", err)
	}

	dm, err := download.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create download manager: %w", err)
	}

	if err := dm.DownloadLoader(vendor, b.Config.Device); err != nil {
		return fmt.Errorf("failed to download loader: %w", err)
	}

	switch vendor {
	case "amlogic":
		return b.writeAmlogicBootloader()
	case "allwinner":
		return b.writeAllwinnerBootloader()
	case "rockchip":
		return b.writeRockchipBootloader()
	default:
		return fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

func (b *Builder) Cleanup() error {
	return os.RemoveAll(b.TempDir)
}
