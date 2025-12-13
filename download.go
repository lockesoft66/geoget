package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func downloadFile(url, destination string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %s while downloading %s", resp.Status, url)
	}

	if err := os.MkdirAll(filepathDir(destination), 0o755); err != nil {
		return fmt.Errorf("create download dir: %w", err)
	}

	out, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("write download: %w", err)
	}

	return nil
}
