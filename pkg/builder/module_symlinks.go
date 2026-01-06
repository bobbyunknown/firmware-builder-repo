package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (b *Builder) createModuleSymlinks(modulesDir, kernelVersion string) error {
	moduleVersionDir := filepath.Join(modulesDir, "lib", "modules", kernelVersion)

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

	for _, koFile := range koFiles {
		linkName := filepath.Join(moduleVersionDir, filepath.Base(koFile))

		if koFile == linkName {
			continue
		}

		os.Remove(linkName)

		relPath, err := filepath.Rel(moduleVersionDir, koFile)
		if err != nil {
			continue
		}

		if err := os.Symlink(relPath, linkName); err != nil {
			fmt.Printf("   Warning: Could not create symlink for %s: %v\n", filepath.Base(koFile), err)
		}
	}

	return nil
}
