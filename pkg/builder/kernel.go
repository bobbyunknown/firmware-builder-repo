package builder

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bobbyunknown/Oh-my-builder/pkg/config"
	"github.com/bobbyunknown/Oh-my-builder/pkg/download"
	"github.com/diskfs/go-diskfs"
)

func (b *Builder) InstallKernel() error {
	fmt.Println("🔧 Installing kernel...")

	dm, err := download.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create download manager: %w", err)
	}

	kernelPath := dm.GetKernelPath(b.Config.Kernel)
	bootDir := filepath.Join(b.TempDir, "boot")
	modulesDir := filepath.Join(b.TempDir, "modules")
	rootDir := filepath.Join(b.TempDir, "device_root")

	if err := os.MkdirAll(bootDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(modulesDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return err
	}

	vendor, err := config.GetDeviceVendor(b.Config.Device)
	if err != nil {
		return err
	}

	bootTar := filepath.Join(kernelPath, fmt.Sprintf("boot-%s.tar.gz", b.Config.Kernel))
	if err := extractTarGz(bootTar, bootDir); err != nil {
		return fmt.Errorf("failed to extract boot: %w", err)
	}
	fmt.Println("   ✓ Extracted boot files")

	dtbTar := filepath.Join(kernelPath, fmt.Sprintf("dtb-%s-%s.tar.gz", vendor, b.Config.Kernel))
	dtbDir := filepath.Join(bootDir, "dtb", vendor)
	if err := os.MkdirAll(dtbDir, 0755); err != nil {
		return err
	}
	if err := extractTarGz(dtbTar, dtbDir); err != nil {
		return fmt.Errorf("failed to extract dtb: %w", err)
	}
	fmt.Println("   ✓ Extracted DTB files")

	modulesTar := filepath.Join(kernelPath, fmt.Sprintf("modules-%s.tar.gz", b.Config.Kernel))
	if err := extractTarGz(modulesTar, modulesDir); err != nil {
		return fmt.Errorf("failed to extract modules: %w", err)
	}
	fmt.Println("   ✓ Extracted kernel modules")

	if err := b.copyModulesToRoot(modulesDir, b.Config.Kernel); err != nil {
		return fmt.Errorf("failed to copy modules to root: %w", err)
	}
	fmt.Println("   ✓ Copied modules to root directory")
	if err := b.extractDeviceFiles(bootDir); err != nil {
		return fmt.Errorf("failed to extract device files: %w", err)
	}

	disk, err := diskfs.Open(b.Config.Output)
	if err != nil {
		return fmt.Errorf("failed to open disk: %w", err)
	}

	fs, err := disk.GetFilesystem(1)
	if err != nil {
		return fmt.Errorf("failed to get boot filesystem: %w", err)
	}

	if err := copyDirToFS(fs, bootDir, "/"); err != nil {
		return fmt.Errorf("failed to copy boot files: %w", err)
	}
	fmt.Println("   ✓ Copied boot files to partition")

	return nil
}

func (b *Builder) extractDeviceFiles(bootDir string) error {
	deviceBootTar := fmt.Sprintf("boot-%s.tar.gz", b.Config.Device)

	dm, err := download.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create download manager: %w", err)
	}

	repo := dm.Config.Repositories["data"]
	cacheDir := repo.CacheDir()
	deviceCacheDir := filepath.Join(cacheDir, "devices", b.Config.Device)
	cachedFile := filepath.Join(deviceCacheDir, deviceBootTar)

	if _, err := os.Stat(cachedFile); os.IsNotExist(err) {
		fmt.Printf("   Downloading device boot files: devices/%s/%s\n", b.Config.Device, deviceBootTar)

		if err := os.MkdirAll(deviceCacheDir, 0755); err != nil {
			return fmt.Errorf("failed to create cache dir: %w", err)
		}

		remotePath := fmt.Sprintf("devices/%s/%s", b.Config.Device, deviceBootTar)
		if err := dm.DownloadFile(remotePath, cachedFile); err != nil {
			fmt.Printf("   Warning: Could not download device boot files: %v\n", err)
			return nil
		}
	} else {
		fmt.Println("   Device boot files already cached")
	}

	if err := extractTarGz(cachedFile, bootDir); err != nil {
		return fmt.Errorf("failed to extract device boot files: %w", err)
	}

	fmt.Println("   ✓ Extracted device boot files")
	return nil
}

func extractTarGz(tarPath, destDir string) error {
	file, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
			if err := os.Chmod(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		}
	}

	return nil
}
