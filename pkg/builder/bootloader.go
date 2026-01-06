package builder

import (
	"fmt"
	"os"

	"github.com/bobbyunknown/Oh-my-builder/pkg/download"
)

func (b *Builder) writeAmlogicBootloader() error {
	dm, err := download.NewManager()
	if err != nil {
		return err
	}

	loaderDir := dm.GetLoaderPath("amlogic", b.Config.Device)
	loaderPath := fmt.Sprintf("%s/%s.bin", loaderDir, b.Config.Device)

	if _, err := os.Stat(loaderPath); os.IsNotExist(err) {
		return fmt.Errorf("loader not found: %s", loaderPath)
	}

	img, err := os.OpenFile(b.Config.Output, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer img.Close()

	loader, err := os.ReadFile(loaderPath)
	if err != nil {
		return fmt.Errorf("failed to read loader: %w", err)
	}

	if len(loader) < 512 {
		return fmt.Errorf("loader file too small: %d bytes", len(loader))
	}

	if _, err := img.WriteAt(loader[:444], 0); err != nil {
		return fmt.Errorf("failed to write first block: %w", err)
	}

	if _, err := img.WriteAt(loader[512:], 512); err != nil {
		return fmt.Errorf("failed to write second block: %w", err)
	}

	fmt.Printf("   ✓ Wrote Amlogic bootloader: %s.bin\n", b.Config.Device)
	return nil
}

func (b *Builder) writeAllwinnerBootloader() error {
	dm, err := download.NewManager()
	if err != nil {
		return err
	}

	loaderDir := dm.GetLoaderPath("allwinner", b.Config.Device)
	loaderPath := fmt.Sprintf("%s/u-boot-sunxi-with-spl-%s.bin", loaderDir, b.Config.Device)

	if _, err := os.Stat(loaderPath); os.IsNotExist(err) {
		return fmt.Errorf("loader not found: %s", loaderPath)
	}

	img, err := os.OpenFile(b.Config.Output, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer img.Close()

	loader, err := os.ReadFile(loaderPath)
	if err != nil {
		return fmt.Errorf("failed to read loader: %w", err)
	}

	if _, err := img.WriteAt(loader, 8192); err != nil {
		return fmt.Errorf("failed to write bootloader: %w", err)
	}

	fmt.Printf("   ✓ Wrote Allwinner bootloader: u-boot-sunxi-with-spl-%s.bin\n", b.Config.Device)

	mainlinePath := fmt.Sprintf("%s/u-boot-mainline-%s.bin", loaderDir, b.Config.Device)
	if mainline, err := os.ReadFile(mainlinePath); err == nil {
		if _, err := img.WriteAt(mainline, 40960); err != nil {
			return fmt.Errorf("failed to write mainline u-boot: %w", err)
		}
		fmt.Printf("   ✓ Wrote mainline u-boot: u-boot-mainline-%s.bin\n", b.Config.Device)
	}

	return nil
}

func (b *Builder) writeRockchipBootloader() error {
	dm, err := download.NewManager()
	if err != nil {
		return err
	}

	loaderDir := dm.GetLoaderPath("rockchip", b.Config.Device)

	img, err := os.OpenFile(b.Config.Output, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer img.Close()

	idbPath := fmt.Sprintf("%s/idbloader-%s.img", loaderDir, b.Config.Device)
	if idb, err := os.ReadFile(idbPath); err == nil {
		if _, err := img.WriteAt(idb, 64*512); err != nil {
			return fmt.Errorf("failed to write idbloader: %w", err)
		}
		fmt.Printf("   ✓ Wrote idbloader: idbloader-%s.img\n", b.Config.Device)
	} else {
		return fmt.Errorf("idbloader not found: %s", idbPath)
	}

	ubootPath := fmt.Sprintf("%s/u-boot-%s.itb", loaderDir, b.Config.Device)
	if uboot, err := os.ReadFile(ubootPath); err == nil {
		if _, err := img.WriteAt(uboot, 16384*512); err != nil {
			return fmt.Errorf("failed to write u-boot: %w", err)
		}
		fmt.Printf("   ✓ Wrote u-boot: u-boot-%s.itb\n", b.Config.Device)
	} else {
		return fmt.Errorf("u-boot not found: %s", ubootPath)
	}

	trustPath := fmt.Sprintf("%s/trust-%s.bin", loaderDir, b.Config.Device)
	if trust, err := os.ReadFile(trustPath); err == nil {
		if _, err := img.WriteAt(trust, 24576*512); err != nil {
			return fmt.Errorf("failed to write trust: %w", err)
		}
		fmt.Printf("   ✓ Wrote trust: trust-%s.bin\n", b.Config.Device)
	}

	return nil
}
