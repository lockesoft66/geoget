package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func extractZip(archivePath, destination string) error {
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return fmt.Errorf("create extraction dir: %w", err)
	}

	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer reader.Close()

	for _, f := range reader.File {
		if err := extractZipFile(f, destination); err != nil {
			return err
		}
	}

	return nil
}

func extractZipFile(f *zip.File, destination string) error {
	// Prevent zip slip by ensuring the final path stays inside destination.
	targetPath := filepath.Join(destination, f.Name)
	if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(destination)+string(filepath.Separator)) {
		return fmt.Errorf("illegal file path in zip: %s", f.Name)
	}

	if f.FileInfo().IsDir() {
		return os.MkdirAll(targetPath, f.Mode())
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("create file dir: %w", err)
	}

	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("open zipped file: %w", err)
	}
	defer rc.Close()

	out, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
	if err != nil {
		return fmt.Errorf("create extracted file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, rc); err != nil {
		return fmt.Errorf("write extracted file: %w", err)
	}

	return nil
}
