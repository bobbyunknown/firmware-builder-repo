package builder

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	ext4fs "github.com/pilat/go-ext4fs"
)

func (b *Builder) writeRootfsWithExt4fs(partition io.ReadWriteSeeker, size int64, rootfsDir string) error {
	tmpImg := filepath.Join(b.TempDir, "rootfs_temp.img")

	img, err := ext4fs.New(
		ext4fs.WithImagePath(tmpImg),
		ext4fs.WithSize(uint64(size)),
	)
	if err != nil {
		return fmt.Errorf("failed to create ext4 filesystem: %w", err)
	}
	defer img.Close()
	defer os.Remove(tmpImg)

	if err := b.copyDirToExt4fs(img, rootfsDir, ext4fs.RootInode); err != nil {
		return fmt.Errorf("failed to copy files: %w", err)
	}

	if err := img.Save(); err != nil {
		return fmt.Errorf("failed to save filesystem: %w", err)
	}

	imgFile, err := os.Open(tmpImg)
	if err != nil {
		return fmt.Errorf("failed to open temp image: %w", err)
	}
	defer imgFile.Close()

	if _, err := io.Copy(partition, imgFile); err != nil {
		return fmt.Errorf("failed to copy to partition: %w", err)
	}

	return nil
}

func (b *Builder) copyDirToExt4fs(img *ext4fs.Image, srcDir string, parentInode uint32) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())

		if entry.IsDir() {
			inode, err := img.CreateDirectory(parentInode, entry.Name(), 0755, 0, 0)
			if err != nil {
				fmt.Printf("   Warning: Could not create directory %s: %v\n", entry.Name(), err)
				continue
			}
			if err := b.copyDirToExt4fs(img, srcPath, inode); err != nil {
				return err
			}
		} else if entry.Type()&os.ModeSymlink != 0 {
			link, err := os.Readlink(srcPath)
			if err != nil {
				continue
			}
			_, err = img.CreateSymlink(parentInode, entry.Name(), link, 0, 0)
			if err != nil {
				fmt.Printf("   Warning: Could not create symlink %s: %v\n", entry.Name(), err)
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				continue
			}

			info, _ := entry.Info()
			mode := uint16(0644)
			if info != nil {
				mode = uint16(info.Mode() & 0777)
			}

			if _, err := img.CreateFile(parentInode, entry.Name(), data, mode, 0, 0); err != nil {
				fmt.Printf("   Warning: Could not create file %s: %v\n", entry.Name(), err)
			}
		}
	}

	return nil
}
