package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractTarGz extracts a .tar.gz file into destDir.
// It handles archives where entries start with "./" prefix.
func ExtractTarGz(archivePath, destDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar read: %w", err)
		}

		name := filepath.Clean(header.Name)
		if name == "." {
			continue
		}

		// Security: prevent path traversal
		if strings.Contains(name, "..") {
			return fmt.Errorf("invalid path in archive: %s", name)
		}

		target := filepath.Join(destDir, name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("mkdir %s: %w", target, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("mkdir parent %s: %w", target, err)
			}
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("create file %s: %w", target, err)
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return fmt.Errorf("write file %s: %w", target, err)
			}
			outFile.Close()
		}
	}

	return nil
}

// InstallExtension downloads and installs an extension.
func InstallExtension(ext *ExtensionInfo, paths *ZedPaths) error {
	destDir := filepath.Join(paths.Installed, ext.ID)

	// Remove existing installation if present
	if _, err := os.Stat(destDir); err == nil {
		fmt.Printf("  removing existing installation...\n")
		if err := os.RemoveAll(destDir); err != nil {
			return fmt.Errorf("remove existing: %w", err)
		}
	}

	// Create temp file for download
	tmpFile, err := os.CreateTemp("", "zed-ext-*.tar.gz")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Download
	url := GetDownloadURL(ext.ID, ext.Version)
	fmt.Printf("  downloading from %s\n", url)
	if err := DownloadFile(url, tmpPath); err != nil {
		return fmt.Errorf("download: %w", err)
	}

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Extract
	fmt.Printf("  extracting to %s\n", destDir)
	if err := ExtractTarGz(tmpPath, destDir); err != nil {
		os.RemoveAll(destDir)
		return fmt.Errorf("extract: %w", err)
	}

	return nil
}

// RemoveExtension removes an installed extension.
func RemoveExtension(id string, paths *ZedPaths) error {
	destDir := filepath.Join(paths.Installed, id)
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		return fmt.Errorf("extension %q is not installed", id)
	}
	return os.RemoveAll(destDir)
}
