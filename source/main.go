package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	geosReleaseURL    = "https://github.com/bluewaysw/pcgeos/releases/download/CI-latest-issue-829/pcgeos-ensemble_nc.zip"
	baseboxReleaseURL = "https://github.com/bluewaysw/pcgeos-basebox/releases/download/CI-latest-issue-13/pcgeos-basebox.zip"
	geosArchiveRoot   = "ensemble"
)

func main() {
	installRoot := parseInstallRoot()

	baseboxDir := filepath.Join(installRoot, "basebox")
	drivecDir := filepath.Join(installRoot, "drivec")
	geosInstallDir := filepath.Join(drivecDir, geosArchiveRoot)

	logger := log.New(os.Stdout, "[geoget] ", 0)

	if err := prepareInstallDirs(installRoot, geosInstallDir, baseboxDir); err != nil {
		fatal(err)
	}

	tempDir, err := os.MkdirTemp("", "geoget-*")
	if err != nil {
		fatal(fmt.Errorf("create temp dir: %w", err))
	}
	defer os.RemoveAll(tempDir)

	geosZip := filepath.Join(tempDir, "pcgeos-ensemble.zip")
	baseboxZip := filepath.Join(tempDir, "pcgeos-basebox.zip")

	logger.Println("Downloading PC/GEOS Ensemble build")
	if err := downloadFile(geosReleaseURL, geosZip); err != nil {
		fatal(fmt.Errorf("download geos: %w", err))
	}

	logger.Println("Downloading Basebox DOSBox-Staging fork")
	if err := downloadFile(baseboxReleaseURL, baseboxZip); err != nil {
		fatal(fmt.Errorf("download basebox: %w", err))
	}

	geosExtractDir := filepath.Join(tempDir, "ensemble")
	baseboxExtractDir := filepath.Join(tempDir, "basebox")

	logger.Println("Extracting Ensemble archive")
	if err := extractZip(geosZip, geosExtractDir); err != nil {
		fatal(fmt.Errorf("extract geos: %w", err))
	}

	logger.Println("Extracting Basebox archive")
	if err := extractZip(baseboxZip, baseboxExtractDir); err != nil {
		fatal(fmt.Errorf("extract basebox: %w", err))
	}

	geosSource, err := resolveGeosArchiveRoot(geosExtractDir)
	if err != nil {
		fatal(err)
	}

	logger.Printf("Installing Ensemble into %s\n", geosInstallDir)
	if err := copyDir(geosSource, geosInstallDir); err != nil {
		fatal(fmt.Errorf("copy geos: %w", err))
	}

	baseboxSource := resolveBaseboxRoot(baseboxExtractDir)
	logger.Printf("Installing Basebox into %s\n", baseboxDir)
	if err := copyDir(baseboxSource, baseboxDir); err != nil {
		fatal(fmt.Errorf("copy basebox: %w", err))
	}

	if err := ensureExecutables(baseboxDir); err != nil {
		fatal(err)
	}

	baseboxBinary, err := detectBaseboxBinary(baseboxDir)
	if err != nil {
		fatal(err)
	}
	logger.Printf("Using Basebox executable: %s (%s)\n", baseboxBinary.relPath, baseboxBinary.arch)

	if err := writeBaseboxConfig(baseboxDir, drivecDir); err != nil {
		fatal(err)
	}

	if err := createLaunchers(installRoot, baseboxBinary.arch); err != nil {
		fatal(err)
	}

	logger.Println("Deployment complete.")
}

func parseInstallRoot() string {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <install-root>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	root := os.Args[1]
	if filepath.IsAbs(root) {
		return filepath.Clean(root)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fatal(fmt.Errorf("resolve home directory: %w", err))
	}

	return filepath.Join(homeDir, root)
}

func prepareInstallDirs(installRoot, geosInstallDir, baseboxDir string) error {
	if installRoot == "" || installRoot == "/" || installRoot == string(filepath.Separator) {
		return fmt.Errorf("refusing to operate on empty install root")
	}

	if err := os.RemoveAll(installRoot); err != nil {
		return fmt.Errorf("remove existing install root: %w", err)
	}

	if err := os.MkdirAll(geosInstallDir, 0o755); err != nil {
		return fmt.Errorf("create geos install dir: %w", err)
	}

	if err := os.MkdirAll(baseboxDir, 0o755); err != nil {
		return fmt.Errorf("create basebox dir: %w", err)
	}

	return nil
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
