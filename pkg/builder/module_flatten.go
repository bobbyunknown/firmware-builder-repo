package builder

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (b *Builder) copyModulesToRoot(modulesDir, kernelVersion string) error {
	moduleVersionDir := filepath.Join(modulesDir, kernelVersion)

	if _, err := os.Stat(moduleVersionDir); os.IsNotExist(err) {
		return nil
	}

	os.Remove(filepath.Join(moduleVersionDir, "build"))
	os.Remove(filepath.Join(moduleVersionDir, "source"))

	koFiles := []string{}
	err := filepath.Walk(moduleVersionDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".ko") {
			koFiles = append(koFiles, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("   Found %d modules to copy...\n", len(koFiles))

	count := 0
	for _, koFile := range koFiles {
		targetName := filepath.Join(moduleVersionDir, filepath.Base(koFile))

		if koFile == targetName {
			continue
		}

		if err := copyFileContent(koFile, targetName); err != nil {
			fmt.Printf("   Warning: Could not copy module %s: %v\n", filepath.Base(koFile), err)
		} else {
			count++
		}
	}

	fmt.Printf("   Copied %d modules to root directory\n", count)
	return nil
}

func copyFileContent(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}
