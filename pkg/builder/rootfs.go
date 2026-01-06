package builder

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bobbyunknown/Oh-my-builder/pkg/config"
	"github.com/bobbyunknown/Oh-my-builder/pkg/download"
	"github.com/ulikunitz/xz"
)

func (b *Builder) InstallRootfs() error {
	fmt.Println("📦 Installing rootfs...")

	dm, err := download.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create download manager: %w", err)
	}

	rootfsPath := dm.GetRootfsPath(b.Config.Rootfs)
	rootfsDir := filepath.Join(b.TempDir, "rootfs")

	if err := os.MkdirAll(rootfsDir, 0755); err != nil {
		return err
	}

	if err := b.extractRootfs(rootfsPath, rootfsDir); err != nil {
		return fmt.Errorf("failed to extract rootfs: %w", err)
	}
	fmt.Println("   ✓ Extracted rootfs")

	modulesDir := filepath.Join(b.TempDir, "modules")
	if _, err := os.Stat(modulesDir); err == nil {
		destModules := filepath.Join(rootfsDir, "lib", "modules")
		if err := copyDir(modulesDir, destModules); err != nil {
			return fmt.Errorf("failed to copy modules: %w", err)
		}
		fmt.Println("   ✓ Copied kernel modules")
	}

	deviceRootDir := filepath.Join(b.TempDir, "device_root")
	if _, err := os.Stat(deviceRootDir); err == nil {
		if err := copyDir(deviceRootDir, rootfsDir); err != nil {
			return fmt.Errorf("failed to copy device files: %w", err)
		}
		fmt.Println("   ✓ Copied device files")
	}

	if err := b.installFirmware(rootfsDir); err != nil {
		return fmt.Errorf("failed to install firmware: %w", err)
	}
	fmt.Println("   ✓ Installed firmware files")

	if err := b.applyTweaks(rootfsDir); err != nil {
		return fmt.Errorf("failed to apply tweaks: %w", err)
	}
	fmt.Println("   ✓ Applied rootfs tweaks")

	fmt.Println("   Writing rootfs to partition...")

	// Open the image file
	f, err := os.OpenFile(b.Config.Output, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer f.Close()

	// Calculate partition offset (partition 2 starts at sector 526336)
	partitionOffset := int64(526336 * 512)
	partitionSize := int64(2127872 * 512)

	// Seek to partition start
	if _, err := f.Seek(partitionOffset, 0); err != nil {
		return fmt.Errorf("failed to seek to partition: %w", err)
	}

	// Create ext4 filesystem on partition using pilat/go-ext4fs
	if err := b.writeRootfsWithExt4fs(f, partitionSize, rootfsDir); err != nil {
		return fmt.Errorf("failed to write rootfs: %w", err)
	}

	fmt.Println("   ✓ Wrote rootfs to partition")
	return nil
}

func (b *Builder) extractRootfs(rootfsPath, destDir string) error {
	ext := strings.ToLower(filepath.Ext(rootfsPath))

	switch ext {
	case ".xz":
		return b.extractXZ(rootfsPath, destDir)
	case ".gz":
		if strings.HasSuffix(rootfsPath, ".tar.gz") {
			return extractTarGz(rootfsPath, destDir)
		}
		return b.extractGZ(rootfsPath, destDir)
	default:
		return fmt.Errorf("unsupported rootfs format: %s", ext)
	}
}

func (b *Builder) extractXZ(xzPath, destDir string) error {
	file, err := os.Open(xzPath)
	if err != nil {
		return err
	}
	defer file.Close()

	xzr, err := xz.NewReader(file)
	if err != nil {
		return err
	}

	imgPath := filepath.Join(b.TempDir, "rootfs.img")
	out, err := os.Create(imgPath)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, xzr); err != nil {
		return err
	}
	out.Close()

	fmt.Println("   Decompressed .xz to .img")
	return b.extractExt4Image(imgPath, destDir)
}

func (b *Builder) extractGZ(gzPath, destDir string) error {
	file, err := os.Open(gzPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	imgPath := filepath.Join(b.TempDir, "rootfs.img")
	out, err := os.Create(imgPath)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, gzr); err != nil {
		return err
	}
	out.Close()

	fmt.Println("   Decompressed .gz to .img")
	return b.extractExt4Image(imgPath, destDir)
}

func (b *Builder) applyTweaks(rootfsDir string) error {
	vendor, err := config.GetDeviceVendor(b.Config.Device)
	if err != nil {
		return err
	}

	switch vendor {
	case "amlogic":
		return b.applyAmlogicTweaks(rootfsDir)
	case "allwinner", "rockchip":
		return b.applyAllwinnerRockchipTweaks(rootfsDir)
	}

	return nil
}

func (b *Builder) applyAmlogicTweaks(rootfsDir string) error {
	pwmFile := filepath.Join(rootfsDir, "etc", "modules.d", "pwm-meson")
	if err := os.MkdirAll(filepath.Dir(pwmFile), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(pwmFile, []byte("pwm_meson\n"), 0644); err != nil {
		return err
	}

	inittab := filepath.Join(rootfsDir, "etc", "inittab")
	if err := replaceInFileOS(inittab, "ttyAMA0", "ttyAML0"); err != nil {
		return err
	}
	if err := replaceInFileOS(inittab, "ttyS0", "tty0"); err != nil {
		return err
	}

	bootScript := filepath.Join(rootfsDir, "etc", "init.d", "boot")
	if err := prependLineBeforeOS(bootScript, "kmodloader", "\tmkdir -p /tmp/upgrade"); err != nil {
		return err
	}

	return nil
}

func (b *Builder) applyAllwinnerRockchipTweaks(rootfsDir string) error {
	inittab := filepath.Join(rootfsDir, "etc", "inittab")
	if err := replaceInFileOS(inittab, "ttyAMA0", "tty1"); err != nil {
		return err
	}
	if err := replaceInFileOS(inittab, "ttyS0", "ttyS2"); err != nil {
		return err
	}

	return b.applyCommonTweaks(rootfsDir)
}

func (b *Builder) applyCommonTweaks(rootfsDir string) error {
	bootScript := filepath.Join(rootfsDir, "etc", "init.d", "boot")
	if err := prependLineBeforeOS(bootScript, "kmodloader", "\tulimit -n 131072"); err != nil {
		return err
	}

	mac80211 := filepath.Join(rootfsDir, "lib", "netifd", "wireless", "mac80211.sh")
	if err := replaceInFileOS(mac80211, "iw ", "ipconfig "); err != nil {
		return err
	}

	return nil
}

func replaceInFileOS(filePath, old, new string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	updated := strings.ReplaceAll(string(content), old, new)
	if updated == string(content) {
		return nil
	}

	return os.WriteFile(filePath, []byte(updated), 0644)
}

func prependLineBeforeOS(filePath, marker, line string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	text := string(content)
	if strings.Contains(text, line) {
		return nil
	}

	idx := strings.Index(text, marker)
	if idx == -1 {
		return nil
	}

	updated := text[:idx] + line + "\n" + text[idx:]
	return os.WriteFile(filePath, []byte(updated), 0644)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

func (b *Builder) installFirmware(rootfsDir string) error {
	dm, err := download.NewManager()
	if err != nil {
		return err
	}

	if err := dm.DownloadFirmware(); err != nil {
		return err
	}

	firmwareSrc := dm.GetFirmwarePath()
	firmwareDst := filepath.Join(rootfsDir, "lib", "firmware")

	if _, err := os.Stat(firmwareSrc); os.IsNotExist(err) {
		return nil
	}

	if err := os.RemoveAll(firmwareDst); err != nil && !os.IsNotExist(err) {
		return err
	}

	return copyDir(firmwareSrc, firmwareDst)
}
