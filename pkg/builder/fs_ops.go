package builder

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/diskfs/go-diskfs/filesystem"
)

func copyDirToFS(fs filesystem.FileSystem, srcDir, destDir string) error {
	return filepath.WalkDir(srcDir, func(localPath string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, localPath)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}

		targetPath := path.Join(destDir, rel)
		if entry.IsDir() {
			return ensureDir(fs, targetPath)
		}

		if err := ensureDir(fs, path.Dir(targetPath)); err != nil {
			return err
		}

		srcFile, err := os.Open(localPath)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := fs.OpenFile(targetPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC)
		if err != nil {
			return err
		}

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			_ = dstFile.Close()
			return err
		}
		if err := dstFile.Close(); err != nil {
			return err
		}

		info, err := entry.Info()
		if err == nil {
			_ = fs.Chmod(targetPath, info.Mode())
		}

		return nil
	})
}

func ensureDir(fs filesystem.FileSystem, dir string) error {
	if dir == "" || dir == "/" || dir == "." {
		return nil
	}
	dir = path.Clean(dir)
	if dir == "/" || dir == "." {
		return nil
	}

	parts := strings.Split(strings.TrimPrefix(dir, "/"), "/")
	current := ""
	for _, part := range parts {
		current = current + "/" + part
		if err := fs.Mkdir(current); err != nil {
			if _, rdErr := fs.ReadDir(current); rdErr == nil {
				continue
			}
			return err
		}
	}
	return nil
}
