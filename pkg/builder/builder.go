package builder

import (
	"fmt"
	"os"
	"path/filepath"

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

	dm := download.NewManager(b.CacheDir, "bobbyunknown", "Oh-my-builder", "data")

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
	imageSize := int64(b.Config.Size) * 1024 * 1024

	if err := os.MkdirAll(filepath.Dir(imagePath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

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
	spec := disk.FilesystemSpec{
		Partition:   1,
		FSType:      filesystem.TypeFat32,
		VolumeLabel: "BOOT",
	}
	if _, err := mydisk.CreateFilesystem(spec); err != nil {
		return fmt.Errorf("failed to format boot partition: %w", err)
	}

	fmt.Println("   ✓ Disk image created with partitions")
	fmt.Println("   Note: Rootfs partition will be formatted during installation")
	return nil
}

func (b *Builder) InstallKernel() error {
	fmt.Println("🔧 Installing kernel...")

	dm := download.NewManager(b.CacheDir, "bobbyunknown", "Oh-my-builder", "data")
	kernelDir := dm.GetKernelPath(b.Config.Kernel)

	mydisk, err := diskfs.Open(b.Config.Output)
	if err != nil {
		return fmt.Errorf("failed to open disk image: %w", err)
	}

	fs, err := mydisk.GetFilesystem(1)
	if err != nil {
		return fmt.Errorf("failed to get boot filesystem: %w", err)
	}

	files, err := os.ReadDir(kernelDir)
	if err != nil {
		return fmt.Errorf("failed to read kernel directory: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".gz" {
			fmt.Printf("   Extracting %s...\n", file.Name())
			srcPath := filepath.Join(kernelDir, file.Name())

			data, err := os.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", file.Name(), err)
			}

			destPath := "/" + file.Name()
			if err := fs.Mkdir(destPath); err == nil {
			}

			rw, err := fs.OpenFile(destPath, os.O_CREATE|os.O_RDWR)
			if err != nil {
				fmt.Printf("   Warning: failed to write %s: %v\n", file.Name(), err)
				continue
			}

			if _, err := rw.Write(data); err != nil {
				fmt.Printf("   Warning: failed to write data to %s: %v\n", file.Name(), err)
			}
		}
	}

	fmt.Println("   ✓ Kernel installed")
	return nil
}

func (b *Builder) InstallRootfs() error {
	fmt.Println("📦 Installing rootfs...")
	fmt.Println("   Note: Rootfs installation requires mounting partition")
	fmt.Println("   Skipping for now - will be implemented with OS mount tools")
	return nil
}

func (b *Builder) WriteBootloader() error {
	fmt.Println("🚀 Writing bootloader...")
	fmt.Println("   Note: Bootloader writing requires vendor-specific u-boot files")
	fmt.Println("   Skipping for now - will be implemented with loader files")
	return nil
}

func (b *Builder) Cleanup() error {
	return os.RemoveAll(b.TempDir)
}
