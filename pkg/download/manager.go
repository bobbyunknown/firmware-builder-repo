package download

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bobbyunknown/Oh-my-builder/pkg/config"
	"github.com/schollz/progressbar/v3"
)

type Manager struct {
	Config *config.Config
	Client *http.Client
}

type FileMetadata struct {
	Name string
	Size int64
	Path string
}

type githubContent struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
	Size int    `json:"size"`
}

func NewManager() (*Manager, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &Manager{
		Config: cfg,
		Client: &http.Client{},
	}, nil
}

func (m *Manager) DownloadFile(remotePath, localPath string) error {
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	repo := m.Config.Repositories["data"]
	owner := m.getOwner(repo.URL)
	repoName := m.getRepo(repo.URL)

	isLFSFile := strings.HasPrefix(remotePath, "kernels/") ||
		strings.HasPrefix(remotePath, "rootfs/") ||
		strings.HasPrefix(remotePath, "devices/")

	var downloadURL string
	if isLFSFile {
		downloadURL = fmt.Sprintf("https://media.githubusercontent.com/media/%s/%s/%s/%s",
			owner, repoName, repo.Branch, remotePath)
	} else {
		downloadURL = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s",
			owner, repoName, repo.Branch, remotePath)
	}

	resp, err := m.Client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		filepath.Base(localPath),
	)

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	fmt.Println()

	return err
}

func (m *Manager) DownloadKernel(version string) error {
	repo := m.Config.Repositories["data"]
	cacheDir := repo.CacheDir()
	kernelDir := filepath.Join(cacheDir, "kernels", version)

	if _, err := os.Stat(kernelDir); err == nil {
		files, err := m.listDirectory(fmt.Sprintf("kernels/%s", version))
		if err == nil {
			allValid := true
			for _, file := range files {
				remotePath := fmt.Sprintf("kernels/%s/%s", version, file)
				localPath := filepath.Join(kernelDir, file)

				valid, err := m.validateCachedFile(localPath, remotePath)
				if err != nil || !valid {
					allValid = false
					break
				}
			}

			if allValid {
				fmt.Printf("Kernel %s already cached and valid\n", version)
				return nil
			}
			fmt.Printf("Kernel %s cache invalid, re-downloading...\n", version)
			os.RemoveAll(kernelDir)
		}
	}

	fmt.Printf("Downloading kernel %s...\n", version)

	files, err := m.listDirectory(fmt.Sprintf("kernels/%s", version))
	if err != nil {
		return fmt.Errorf("failed to list kernel files: %w", err)
	}

	for _, file := range files {
		remotePath := fmt.Sprintf("kernels/%s/%s", version, file)
		localPath := filepath.Join(kernelDir, file)

		if err := m.DownloadFile(remotePath, localPath); err != nil {
			fmt.Printf("Warning: failed to download %s: %v\n", file, err)
		}
	}

	return nil
}

func (m *Manager) DownloadRootfs(name string) error {
	repo := m.Config.Repositories["data"]
	cacheDir := repo.CacheDir()
	localPath := filepath.Join(cacheDir, "rootfs", name)
	remotePath := fmt.Sprintf("rootfs/%s", name)

	if _, err := os.Stat(localPath); err == nil {
		valid, err := m.validateCachedFile(localPath, remotePath)
		if err == nil && valid {
			fmt.Printf("Rootfs %s already cached and valid\n", name)
			return nil
		}
		fmt.Printf("Rootfs %s cache invalid, re-downloading...\n", name)
		os.Remove(localPath)
	}

	fmt.Printf("Downloading rootfs %s...\n", name)
	return m.DownloadFile(remotePath, localPath)
}

func (m *Manager) GetKernelPath(version string) string {
	repo := m.Config.Repositories["data"]
	cacheDir := repo.CacheDir()
	return filepath.Join(cacheDir, "kernels", version)
}

func (m *Manager) GetRootfsPath(name string) string {
	repo := m.Config.Repositories["data"]
	cacheDir := repo.CacheDir()
	return filepath.Join(cacheDir, "rootfs", name)
}

func (m *Manager) DownloadLoader(vendor, device string) error {
	repo := m.Config.Repositories["data"]
	cacheDir := repo.CacheDir()
	loaderDir := filepath.Join(cacheDir, "loader", vendor)

	if _, err := os.Stat(loaderDir); err == nil {
		fmt.Printf("Loader files for %s already cached\n", vendor)
		return nil
	}

	fmt.Printf("Downloading loader folder for %s from GitHub...\n", vendor)

	owner := m.getOwner(repo.URL)
	repoName := m.getRepo(repo.URL)

	archiveURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/zipball/%s",
		owner, repoName, repo.Branch)

	req, err := http.NewRequest("GET", archiveURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := m.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download archive: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download archive: HTTP %d", resp.StatusCode)
	}

	tempZip := filepath.Join(cacheDir, "temp_loader.zip")
	out, err := os.Create(tempZip)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		fmt.Sprintf("loader-%s.zip", vendor),
	)

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	out.Close()
	fmt.Println()

	if err != nil {
		os.Remove(tempZip)
		return fmt.Errorf("failed to download: %w", err)
	}

	fmt.Println("Extracting loader files...")
	tempExtract := filepath.Join(cacheDir, "temp_loader_extract")
	if err := m.extractZip(tempZip, tempExtract); err != nil {
		os.Remove(tempZip)
		return fmt.Errorf("failed to extract: %w", err)
	}

	entries, err := os.ReadDir(tempExtract)
	if err != nil {
		os.RemoveAll(tempExtract)
		os.Remove(tempZip)
		return fmt.Errorf("failed to read extracted dir: %w", err)
	}

	if len(entries) == 0 {
		os.RemoveAll(tempExtract)
		os.Remove(tempZip)
		return fmt.Errorf("no files extracted")
	}

	repoRoot := filepath.Join(tempExtract, entries[0].Name())
	loaderSrc := filepath.Join(repoRoot, "loader", vendor)

	if _, err := os.Stat(loaderSrc); os.IsNotExist(err) {
		os.RemoveAll(tempExtract)
		os.Remove(tempZip)
		return fmt.Errorf("loader/%s folder not found in archive", vendor)
	}

	if err := os.MkdirAll(filepath.Dir(loaderDir), 0755); err != nil {
		os.RemoveAll(tempExtract)
		os.Remove(tempZip)
		return err
	}

	if err := os.Rename(loaderSrc, loaderDir); err != nil {
		os.RemoveAll(tempExtract)
		os.Remove(tempZip)
		return fmt.Errorf("failed to move loader: %w", err)
	}

	os.RemoveAll(tempExtract)
	os.Remove(tempZip)

	fmt.Printf("✓ Loader for %s downloaded and extracted\n", vendor)
	return nil
}

func (m *Manager) GetLoaderPath(vendor, device string) string {
	repo := m.Config.Repositories["data"]
	cacheDir := repo.CacheDir()
	return filepath.Join(cacheDir, "loader", vendor)
}

func (m *Manager) DownloadFirmware() error {
	repo := m.Config.Repositories["data"]
	cacheDir := repo.CacheDir()
	firmwareDir := filepath.Join(cacheDir, "firmware")

	if _, err := os.Stat(firmwareDir); err == nil {
		fmt.Println("Firmware files already cached")
		return nil
	}

	fmt.Println("Downloading firmware folder from GitHub...")

	owner := m.getOwner(repo.URL)
	repoName := m.getRepo(repo.URL)

	archiveURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/zipball/%s",
		owner, repoName, repo.Branch)

	req, err := http.NewRequest("GET", archiveURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := m.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download archive: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download archive: HTTP %d", resp.StatusCode)
	}

	tempZip := filepath.Join(cacheDir, "temp_firmware.zip")
	out, err := os.Create(tempZip)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"firmware.zip",
	)

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	out.Close()
	fmt.Println()

	if err != nil {
		os.Remove(tempZip)
		return fmt.Errorf("failed to download: %w", err)
	}

	fmt.Println("Extracting firmware files...")
	tempExtract := filepath.Join(cacheDir, "temp_extract")
	if err := m.extractZip(tempZip, tempExtract); err != nil {
		os.Remove(tempZip)
		return fmt.Errorf("failed to extract: %w", err)
	}

	entries, err := os.ReadDir(tempExtract)
	if err != nil {
		os.RemoveAll(tempExtract)
		os.Remove(tempZip)
		return fmt.Errorf("failed to read extracted dir: %w", err)
	}

	if len(entries) == 0 {
		os.RemoveAll(tempExtract)
		os.Remove(tempZip)
		return fmt.Errorf("no files extracted")
	}

	repoRoot := filepath.Join(tempExtract, entries[0].Name())
	firmwareSrc := filepath.Join(repoRoot, "firmware")

	if _, err := os.Stat(firmwareSrc); os.IsNotExist(err) {
		os.RemoveAll(tempExtract)
		os.Remove(tempZip)
		return fmt.Errorf("firmware folder not found in archive")
	}

	if err := os.Rename(firmwareSrc, firmwareDir); err != nil {
		os.RemoveAll(tempExtract)
		os.Remove(tempZip)
		return fmt.Errorf("failed to move firmware: %w", err)
	}

	os.RemoveAll(tempExtract)
	os.Remove(tempZip)

	fmt.Println("✓ Firmware downloaded and extracted")
	return nil
}

func (m *Manager) GetFirmwarePath() string {
	repo := m.Config.Repositories["data"]
	cacheDir := repo.CacheDir()
	return filepath.Join(cacheDir, "firmware")
}

func (m *Manager) ValidateCache() error {
	repo := m.Config.Repositories["data"]
	cacheDir := repo.CacheDir()

	fmt.Println("\n🔍 Validating cached files...")

	validCount := 0
	invalidCount := 0

	kernelsDir := filepath.Join(cacheDir, "kernels")
	if _, err := os.Stat(kernelsDir); err == nil {
		entries, _ := os.ReadDir(kernelsDir)
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			version := entry.Name()
			files, err := m.listDirectory(fmt.Sprintf("kernels/%s", version))
			if err != nil {
				continue
			}

			allValid := true
			for _, file := range files {
				remotePath := fmt.Sprintf("kernels/%s/%s", version, file)
				localPath := filepath.Join(kernelsDir, version, file)

				valid, err := m.validateCachedFile(localPath, remotePath)
				if err != nil || !valid {
					allValid = false
					invalidCount++
				} else {
					validCount++
				}
			}

			if allValid {
				fmt.Printf("   ✓ Kernel %s: valid\n", version)
			} else {
				fmt.Printf("   ✗ Kernel %s: invalid (will re-download on next use)\n", version)
			}
		}
	}

	rootfsDir := filepath.Join(cacheDir, "rootfs")
	if _, err := os.Stat(rootfsDir); err == nil {
		entries, _ := os.ReadDir(rootfsDir)
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			remotePath := fmt.Sprintf("rootfs/%s", name)
			localPath := filepath.Join(rootfsDir, name)

			valid, err := m.validateCachedFile(localPath, remotePath)
			if err != nil || !valid {
				fmt.Printf("   ✗ Rootfs %s: invalid\n", name)
				invalidCount++
			} else {
				fmt.Printf("   ✓ Rootfs %s: valid\n", name)
				validCount++
			}
		}
	}

	loaderArchives := []string{"amlogic.tar.gz", "allwinner.tar.gz", "rockchip.tar.gz"}
	for _, archive := range loaderArchives {
		archivePath := filepath.Join(cacheDir, "loader", archive)
		if _, err := os.Stat(archivePath); err == nil {
			remotePath := fmt.Sprintf("loader/%s", archive)
			valid, err := m.validateCachedFile(archivePath, remotePath)
			if err != nil || !valid {
				fmt.Printf("   ✗ Loader %s: invalid\n", archive)
				invalidCount++
			} else {
				fmt.Printf("   ✓ Loader %s: valid\n", archive)
				validCount++
			}
		}
	}

	firmwareArchive := filepath.Join(cacheDir, "firmware.tar.gz")
	if _, err := os.Stat(firmwareArchive); err == nil {
		valid, err := m.validateCachedFile(firmwareArchive, "firmware.tar.gz")
		if err != nil || !valid {
			fmt.Println("   ✗ Firmware: invalid")
			invalidCount++
		} else {
			fmt.Println("   ✓ Firmware: valid")
			validCount++
		}
	}

	fmt.Printf("\n📊 Cache validation summary:\n")
	fmt.Printf("   Valid files: %d\n", validCount)
	fmt.Printf("   Invalid files: %d\n", invalidCount)

	if invalidCount > 0 {
		fmt.Println("\n💡 Invalid files will be automatically re-downloaded on next use")
	}

	return nil
}

func (m *Manager) listDirectoryRecursive(path string) ([]string, error) {
	var allFiles []string

	items, err := m.listDirectoryItems(path)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		itemPath := fmt.Sprintf("%s/%s", path, item.Name)

		if item.Type == "file" {
			allFiles = append(allFiles, itemPath)
		} else if item.Type == "dir" {
			subFiles, err := m.listDirectoryRecursive(itemPath)
			if err != nil {
				fmt.Printf("Warning: failed to scan directory %s: %v\n", itemPath, err)
				continue
			}
			allFiles = append(allFiles, subFiles...)
		}
	}

	return allFiles, nil
}

func (m *Manager) listDirectory(path string) ([]string, error) {
	repo := m.Config.Repositories["data"]
	owner := m.getOwner(repo.URL)
	repoName := m.getRepo(repo.URL)

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
		owner, repoName, path, repo.Branch)

	resp, err := m.Client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list directory: HTTP %d", resp.StatusCode)
	}

	var items []githubContent
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var files []string
	for _, item := range items {
		if item.Type == "file" {
			files = append(files, item.Name)
		}
	}

	return files, nil
}

func (m *Manager) listDirectoryItems(path string) ([]githubContent, error) {
	repo := m.Config.Repositories["data"]
	owner := m.getOwner(repo.URL)
	repoName := m.getRepo(repo.URL)

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
		owner, repoName, path, repo.Branch)

	resp, err := m.Client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list directory: HTTP %d", resp.StatusCode)
	}

	var items []githubContent
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return items, nil
}

func (m *Manager) getOwner(url string) string {
	parts := filepath.Base(filepath.Dir(url))
	return parts
}

func (m *Manager) getRepo(url string) string {
	return filepath.Base(url)
}

func (m *Manager) getRemoteFileMetadata(remotePath string) (*FileMetadata, error) {
	repo := m.Config.Repositories["data"]
	owner := m.getOwner(repo.URL)
	repoName := m.getRepo(repo.URL)

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
		owner, repoName, remotePath, repo.Branch)

	resp, err := m.Client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get file metadata: HTTP %d", resp.StatusCode)
	}

	var content githubContent
	if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	return &FileMetadata{
		Name: content.Name,
		Size: int64(content.Size),
		Path: content.Path,
	}, nil
}

func (m *Manager) validateCachedFile(localPath, remotePath string) (bool, error) {
	fileInfo, err := os.Stat(localPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to stat file: %w", err)
	}

	metadata, err := m.getRemoteFileMetadata(remotePath)
	if err != nil {
		return false, fmt.Errorf("failed to get remote metadata: %w", err)
	}

	if fileInfo.Size() != metadata.Size {
		return false, nil
	}

	return true, nil
}

func (m *Manager) extractTarGz(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		target := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}
			outFile, err := os.Create(target)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to write file: %w", err)
			}
			outFile.Close()
		}
	}

	return nil
}

func (m *Manager) extractZip(archivePath, destDir string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		target := filepath.Join(destDir, f.Name)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}

		outFile, err := os.Create(target)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return fmt.Errorf("failed to open file in zip: %w", err)
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()

		if err != nil {
			return fmt.Errorf("failed to extract file: %w", err)
		}
	}

	return nil
}
