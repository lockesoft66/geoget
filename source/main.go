package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultGeosReleaseTag    = "CI-latest"
	defaultBaseboxReleaseTag = "CI-latest"
	geosReleaseBaseURL       = "https://github.com/bluewaysw/pcgeos/releases/download"
	baseboxReleaseBaseURL    = "https://github.com/bluewaysw/pcgeos-basebox/releases/download"
	geosArchiveName          = "pcgeos-ensemble_nc.zip"
	baseboxArchiveName       = "pcgeos-basebox.zip"
	geosArchiveRoot          = "ensemble"
)

func main() {
	installRoot, force, geosTag, baseboxTag, err := parseInstallRootAndFlags()
	if err != nil {
		fatal(err)
	}

	baseboxDir := filepath.Join(installRoot, "basebox")
	drivecDir := filepath.Join(installRoot, "drivec")
	geosInstallDir := filepath.Join(drivecDir, geosArchiveRoot)

	logger := log.New(os.Stdout, "[geoget] ", 0)

	if err := prepareInstallRoot(installRoot, force); err != nil {
		fatal(err)
	}

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
	if err := downloadFile(buildGeosReleaseURL(geosTag), geosZip); err != nil {
		fatal(fmt.Errorf("download geos: %w", err))
	}

	logger.Println("Downloading Basebox DOSBox-Staging fork")
	if err := downloadFile(buildBaseboxReleaseURL(baseboxTag), baseboxZip); err != nil {
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

func parseInstallRootAndFlags() (string, bool, string, string, error) {
	var force bool
	var help bool
	var geosIssue string
	var baseboxIssue string

	flag.BoolVar(&force, "force", false, "overwrite existing installation without prompt")
	flag.BoolVar(&force, "f", false, "overwrite existing installation without prompt")
	flag.BoolVar(&help, "help", false, "show this help message")
	flag.BoolVar(&help, "h", false, "show this help message")
	flag.StringVar(&geosIssue, "geos", "", "GEOS issue number (e.g., 829 or #829)")
	flag.StringVar(&geosIssue, "g", "", "GEOS issue number (e.g., 829 or #829)")
	flag.StringVar(&baseboxIssue, "basebox", "", "Basebox issue number (e.g., 13 or #13)")
	flag.StringVar(&baseboxIssue, "b", "", "Basebox issue number (e.g., 13 or #13)")

	flag.Usage = printUsage
	flag.Parse()

	if help {
		printUsage()
		os.Exit(0)
	}

	geosTag, err := resolveIssueTag(geosIssue, defaultGeosReleaseTag, "GEOS")
	if err != nil {
		return "", false, "", "", err
	}

	baseboxTag, err := resolveIssueTag(baseboxIssue, defaultBaseboxReleaseTag, "Basebox")
	if err != nil {
		return "", false, "", "", err
	}

	root := "geospc"
	if arg := flag.Arg(0); arg != "" {
		root = arg
	}

	if filepath.IsAbs(root) {
		return filepath.Clean(root), force, geosTag, baseboxTag, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", false, "", "", fmt.Errorf("resolve home directory: %w", err)
	}

	return filepath.Join(homeDir, root), force, geosTag, baseboxTag, nil
}

func printUsage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] [install_root]\n", filepath.Base(os.Args[0]))
	fmt.Fprintln(flag.CommandLine.Output())
	fmt.Fprintln(flag.CommandLine.Output(), "Options:")
	fmt.Fprintln(flag.CommandLine.Output(), "  -f, --force            overwrite existing installation without prompt")
	fmt.Fprintln(flag.CommandLine.Output(), "  -g, --geos <issue>     use CI-latest-<issue> for GEOS downloads (accepts 829 or #829)")
	fmt.Fprintln(flag.CommandLine.Output(), "  -b, --basebox <issue>  use CI-latest-<issue> for Basebox downloads (accepts 13 or #13)")
	fmt.Fprintln(flag.CommandLine.Output(), "  -h, --help             show this help message")
	fmt.Fprintln(flag.CommandLine.Output())
	fmt.Fprintln(flag.CommandLine.Output(), "Arguments:")
	fmt.Fprintln(flag.CommandLine.Output(), "  install_root  optional install root; defaults to geospc (absolute roots cleaned, relative roots resolved under home)")
	fmt.Fprintln(flag.CommandLine.Output())
	fmt.Fprintln(flag.CommandLine.Output(), "Defaults:")
	fmt.Fprintln(flag.CommandLine.Output(), "  If no issue flags are provided, CI-latest is used.")
}

func prepareInstallRoot(installRoot string, force bool) error {
	if installRoot == "" || installRoot == "/" || installRoot == string(filepath.Separator) {
		return fmt.Errorf("refusing to operate on empty install root")
	}

	if _, err := os.Stat(installRoot); err == nil {
		if !force {
			confirmed, confirmErr := confirmOverwrite()
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

func prepareInstallDirs(installRoot, geosInstallDir, baseboxDir string) error {
	if installRoot == "" || installRoot == "/" || installRoot == string(filepath.Separator) {
		return fmt.Errorf("refusing to operate on empty install root")
	}

	if err := os.MkdirAll(geosInstallDir, 0o755); err != nil {
		return fmt.Errorf("create geos install dir: %w", err)
	}

	if err := os.MkdirAll(baseboxDir, 0o755); err != nil {
		return fmt.Errorf("create basebox dir: %w", err)
	}

	return nil
}

func confirmOverwrite() (bool, error) {
	fmt.Print("Install root exists, are you really sure you want to overwrite it? [y/N]: ")

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

func buildGeosReleaseURL(tag string) string {
	return fmt.Sprintf("%s/%s/%s", geosReleaseBaseURL, tag, geosArchiveName)
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

func isNumeric(input string) bool {
	for _, ch := range input {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return input != ""
}
