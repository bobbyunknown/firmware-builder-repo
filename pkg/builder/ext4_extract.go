package builder

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	ext4 "github.com/masahiro331/go-ext4-filesystem/ext4"
)

func (b *Builder) extractExt4Image(imgPath, destDir string) error {
	fmt.Println("   Extracting ext4 image contents...")

	f, err := os.Open(imgPath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer f.Close()

	filesize, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to seek start: %w", err)
	}

	filesystem, err := ext4.NewFS(*io.NewSectionReader(f, 0, filesize), nil)
	if err != nil {
		return fmt.Errorf("failed to create ext4 filesystem: %w", err)
	}

	err = fs.WalkDir(filesystem, "/", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, path)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		srcFile, err := filesystem.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", path, err)
		}
		defer srcFile.Close()

		dstFile, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create %s: %w", destPath, err)
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return fmt.Errorf("failed to copy %s: %w", path, err)
		}

		os.Chmod(destPath, info.Mode())

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk ext4: %w", err)
	}

	fmt.Println("   ✓ Extracted ext4 image")
	return nil
}
