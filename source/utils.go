package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func prepareInstallRoot(installRoot string, force bool) error {
	if installRoot == "" || installRoot == "/" || installRoot == string(filepath.Separator) {
		return fmt.Errorf("refusing to operate on empty install root")
	}

	if _, err := os.Stat(installRoot); err == nil {
		if !force {
			confirmed, confirmErr := confirmOverwrite(installRoot)
			if confirmErr != nil {
				return confirmErr
			}
			if !confirmed {
				return errors.New("installation aborted by user")
			}
		}

		if err := os.RemoveAll(installRoot); err != nil {
			return fmt.Errorf("remove existing install root: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check install root: %w", err)
	}

	return nil
}

func prepareInstallDirs(installRoot, drivecDir, baseboxDir string) error {
	if installRoot == "" || installRoot == "/" || installRoot == string(filepath.Separator) {
		return fmt.Errorf("refusing to operate on empty install root")
	}

	if err := os.MkdirAll(drivecDir, 0o755); err != nil {
		return fmt.Errorf("create \"drive c\" dir: %w", err)
	}

	if err := os.MkdirAll(baseboxDir, 0o755); err != nil {
		return fmt.Errorf("create basebox dir: %w", err)
	}

	return nil
}

func confirmOverwrite(installRoot string) (bool, error) {
	fmt.Printf("Install root '%s' exists, are you really sure you want to overwrite it? [y/n]: ", installRoot)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("read confirmation: %w", err)
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes", nil
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}

func buildGeosReleaseURL(tag string, geosLang string) string {
	return fmt.Sprintf("%s/%s/%s%s.zip", geosReleaseBaseURL, tag, geosArchiveName, geosLang)
}

func buildBaseboxReleaseURL(tag string) string {
	return fmt.Sprintf("%s/%s/%s", baseboxReleaseBaseURL, tag, baseboxArchiveName)
}

func resolveIssueTag(input, defaultTag, label string) (string, error) {
	issue := strings.TrimSpace(input)
	if issue == "" {
		return defaultTag, nil
	}

	issue = strings.TrimPrefix(issue, "#")
	if issue == "" {
		return "", fmt.Errorf("%s issue number cannot be empty", label)
	}

	if !isNumeric(issue) {
		return "", fmt.Errorf("%s issue number must be numeric: %q", label, input)
	}

	return fmt.Sprintf("CI-latest-issue-%s", issue), nil
}

func resolveBaseboxRoot(baseDir string) string {
	candidate := filepath.Join(baseDir, "pcgeos-basebox")
	if exists(candidate) {
		return candidate
	}
	return baseDir
}

func resolveGeosLoaderDir(baseDir string) (string, error) {
	var found string

	err := filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		if d.Name() == "loader.exe" {
			found = filepath.Base(filepath.Dir(path))
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("search LOADER.EXE: %w", err)
	}

	if found == "" {
		return found, fmt.Errorf("LOADER.EXE not found: %w", err)
	}

	return found, nil
}

func isNumeric(input string) bool {
	for _, ch := range input {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return input != ""
}

func filepathDir(path string) string {
	dir := filepath.Dir(path)
	if dir == "." {
		return ""
	}
	return dir
}

func filepathBase(path string) string {
	return filepath.Base(path)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
