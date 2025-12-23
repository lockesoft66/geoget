package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	defaultGeosReleaseTag    = "CI-latest"
	defaultBaseboxReleaseTag = "CI-latest"
	geosReleaseBaseURL       = "https://github.com/bluewaysw/pcgeos/releases/download"
	baseboxReleaseBaseURL    = "https://github.com/bluewaysw/pcgeos-basebox/releases/download"
	geosArchiveName          = "pcgeos-ensemble_"
	baseboxArchiveName       = "pcgeos-basebox.zip"
	geosArchiveRoot          = "ensemble"
)

func main() {

	/*
		Prepare
	*/

	installRoot, force, geosTag, baseboxTag, geosLang, err := parseInstallRootAndFlags()
	if err != nil {
		fatal(err)
	}

	baseboxDir := filepath.Join(installRoot, "basebox")
	drivecDir := filepath.Join(installRoot, "drivec")

	logger := log.New(os.Stdout, "[geoget] ", 0)

	if err := prepareInstallRoot(installRoot, force); err != nil {
		fatal(err)
	}

	if err := prepareInstallDirs(installRoot, drivecDir, baseboxDir); err != nil {
		fatal(err)
	}

	tempDir, err := os.MkdirTemp("", "geoget-*")
	if err != nil {
		fatal(fmt.Errorf("create temp dir: %w", err))
	}
	defer os.RemoveAll(tempDir)

	/*
		Download
	*/

	geosZip := filepath.Join(tempDir, "pcgeos-ensemble.zip")
	baseboxZip := filepath.Join(tempDir, "pcgeos-basebox.zip")

	logger.Println("Downloading PC/GEOS Ensemble build")
	if err := downloadFile(buildGeosReleaseURL(geosTag, geosLang), geosZip); err != nil {
		fatal(fmt.Errorf("download geos: %w", err))
	}

	logger.Println("Downloading Basebox")
	if err := downloadFile(buildBaseboxReleaseURL(baseboxTag), baseboxZip); err != nil {
		fatal(fmt.Errorf("download basebox: %w", err))
	}

	/*
		Extract
	*/

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

	/*
		Copy
	*/

	logger.Printf("Installing Ensemble into %s\n", drivecDir)
	if err := copyDir(geosExtractDir, drivecDir); err != nil {
		fatal(fmt.Errorf("copy geos: %w", err))
	}

	baseboxSource := resolveBaseboxRoot(baseboxExtractDir)
	logger.Printf("Installing Basebox into %s\n", baseboxDir)
	if err := copyDir(baseboxSource, baseboxDir); err != nil {
		fatal(fmt.Errorf("copy basebox: %w", err))
	}

	/*
		Ensure excecutables
	*/

	if err := ensureExecutables(baseboxDir); err != nil {
		fatal(err)
	}

	baseboxBinary, err := detectBaseboxBinary(baseboxDir)
	if err != nil {
		fatal(err)
	}
	logger.Printf("Using Basebox executable: %s (%s)\n", baseboxBinary.relPath, baseboxBinary.arch)

	/*
		Write config, create Launchers
	*/
	if err := writeBaseboxConfig(baseboxDir, drivecDir); err != nil {
		fatal(err)
	}

	if err := createLaunchers(installRoot, baseboxBinary.arch); err != nil {
		fatal(err)
	}

	logger.Println("Deployment complete.")
}

func parseInstallRootAndFlags() (string, bool, string, string, string, error) {
	var force bool
	var help bool
	var geosIssue string
	var baseboxIssue string
	var lang string

	flag.BoolVar(&force, "force", false, "overwrite existing installation without prompt")
	flag.BoolVar(&force, "f", false, "overwrite existing installation without prompt")
	flag.BoolVar(&help, "help", false, "show this help message")
	flag.BoolVar(&help, "h", false, "show this help message")
	flag.StringVar(&geosIssue, "geos", "", "GEOS issue number (e.g., 829 or #829)")
	flag.StringVar(&geosIssue, "g", "", "GEOS issue number (e.g., 829 or #829)")
	flag.StringVar(&baseboxIssue, "basebox", "", "Basebox issue number (e.g., 13 or #13)")
	flag.StringVar(&baseboxIssue, "b", "", "Basebox issue number (e.g., 13 or #13)")
	flag.StringVar(&lang, "lang", "", "non-english GEOS language to install (\"gr\")")
	flag.StringVar(&lang, "l", "", "non-english GEOS language to install (\"gr\")")

	flag.Usage = printUsage
	flag.Parse()

	if help {
		printUsage()
		os.Exit(0)
	}

	geosTag, err := resolveIssueTag(geosIssue, defaultGeosReleaseTag, "GEOS")
	if err != nil {
		return "", false, "", "", "", err
	}

	baseboxTag, err := resolveIssueTag(baseboxIssue, defaultBaseboxReleaseTag, "Basebox")
	if err != nil {
		return "", false, "", "", "", err
	}

	if lang != "gr" {
		lang = "nc"
	} else {
		lang = "german"
	}

	root := "geospc"
	if arg := flag.Arg(0); arg != "" {
		root = arg
	}

	if filepath.IsAbs(root) {
		return filepath.Clean(root), force, geosTag, baseboxTag, lang, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", false, "", "", "", fmt.Errorf("resolve home directory: %w", err)
	}

	return filepath.Join(homeDir, root), force, geosTag, baseboxTag, lang, nil
}

func printUsage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] [install_root]\n", filepath.Base(os.Args[0]))
	fmt.Fprintln(flag.CommandLine.Output())
	fmt.Fprintln(flag.CommandLine.Output(), "Options:")
	fmt.Fprintln(flag.CommandLine.Output(), "  -f, --force            overwrite existing installation without prompt")
	fmt.Fprintln(flag.CommandLine.Output(), "  -g, --geos <issue>     use CI-latest-<issue> for GEOS downloads (accepts 829 or #829)")
	fmt.Fprintln(flag.CommandLine.Output(), "  -b, --basebox <issue>  use CI-latest-<issue> for Basebox downloads (accepts 13 or #13)")
	fmt.Fprintln(flag.CommandLine.Output(), "  -h, --help             show this help message")
	fmt.Fprintln(flag.CommandLine.Output(), "  -l, --lang <lang>      non-english GEOS language to install (only \"gr\" supported for now)")
	fmt.Fprintln(flag.CommandLine.Output())
	fmt.Fprintln(flag.CommandLine.Output(), "Arguments:")
	fmt.Fprintln(flag.CommandLine.Output(), "  install_root           optional install root; defaults to \"geospc\" under home")
	fmt.Fprintln(flag.CommandLine.Output())
	fmt.Fprintln(flag.CommandLine.Output(), "Defaults:")
	fmt.Fprintln(flag.CommandLine.Output(), "  If no issue flags are provided, CI-latest is used.")
}

