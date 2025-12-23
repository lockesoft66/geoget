package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type baseboxBinary struct {
	arch    string
	relPath string
}

type launcherTemplate struct {
	templateName string
	outputName   string
	executable   bool
}

var (
	baseboxBinaryPaths = map[string]string{
		"l64":   filepath.Join("binl64", "basebox"),
		"mac":   filepath.Join("binmac", "basebox"),
		"nt":    filepath.Join("binnt", "basebox.exe"),
		"nt64":  filepath.Join("binnt64", "basebox.exe"),
		"rpi64": filepath.Join("binrpi64", "basebox"),
	}
)

func createLaunchers(installRoot, arch string) error {
	launchers, err := launcherTemplatesForArch(arch)
	if err != nil {
		return err
	}

	for _, launcher := range launchers {
		dest := filepath.Join(installRoot, launcher.outputName)

		content, err := templateFS.ReadFile("templ/" + launcher.templateName)
		if err != nil {
			return fmt.Errorf("read launcher template %s: %w", launcher.templateName, err)
		}

		if err := os.WriteFile(dest, content, 0o755); err != nil {
			return fmt.Errorf("write launcher %s: %w", dest, err)
		}

		if launcher.executable && (runtime.GOOS == "linux" || runtime.GOOS == "darwin") {
			if err := os.Chmod(dest, 0o755); err != nil {
				return fmt.Errorf("chmod launcher %s: %w", dest, err)
			}
		}
	}

	return nil
}

func launcherTemplatesForArch(arch string) ([]launcherTemplate, error) {
	switch arch {
	case "l64", "mac", "rpi64":
		return []launcherTemplate{
			{
				templateName: fmt.Sprintf("ensemble.%s.sh", arch),
				outputName:   "ensemble.sh",
				executable:   true,
			},
		}, nil
	case "nt":
		return []launcherTemplate{
			{templateName: "ensemble.nt.cmd", outputName: "ensemble.cmd"},
		}, nil
	case "nt64":
		return []launcherTemplate{
			{templateName: "ensemble.nt64.cmd", outputName: "ensemble.cmd"},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported basebox architecture %q", arch)
	}
}

func writeBaseboxConfig(baseboxDir, drivecDir string) error {

	var loaderDir string
	var config string

	data, err := templateFS.ReadFile("templ/basebox.conf")
	if err != nil {
		return fmt.Errorf("read basebox template: %w", err)
	}

	config = string(data);

	loaderDir, err = resolveGeosLoaderDir(drivecDir)
	if err == nil {
		config = strings.ReplaceAll(config, "{{LOADER_DIR}}", loaderDir)
	}

	hostPath := filepath.Clean(drivecDir)
	config = strings.ReplaceAll(config, "{{HOST_PATH}}", hostPath)

	dest := filepath.Join(baseboxDir, "basebox.conf")
	if err := os.WriteFile(dest, []byte(config), 0o644); err != nil {
		return fmt.Errorf("write basebox.conf: %w", err)
	}

	return nil
}

func ensureExecutables(baseboxDir string) error {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		for _, relPath := range baseboxBinaryPaths {
			exe := filepath.Join(baseboxDir, relPath)
			if exists(exe) {
				if err := os.Chmod(exe, 0o755); err != nil {
					return fmt.Errorf("mark executable %s: %w", exe, err)
				}
			}
		}
	}

	// Also mark shell launchers shipped inside the Basebox archive.
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		_ = filepath.Walk(baseboxDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}
			if strings.HasSuffix(info.Name(), ".sh") {
				_ = os.Chmod(path, 0o755)
			}
			return nil
		})
	}

	return nil
}

func detectBaseboxBinary(baseboxDir string) (baseboxBinary, error) {
	for _, arch := range orderedBaseboxArchs() {
		if relPath, ok := binaryPathForArch(baseboxDir, arch); ok {
			return baseboxBinary{arch: arch, relPath: relPath}, nil
		}
	}

	return baseboxBinary{}, fmt.Errorf("unable to locate the Basebox executable inside %s", baseboxDir)
}

func orderedBaseboxArchs() []string {
	var archs []string

	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			archs = []string{"l64"}
		case "arm64":
			archs = []string{"rpi64"}
		}
	case "darwin":
		if runtime.GOARCH == "amd64" {
			archs = []string{"mac"}
		}
	case "windows":
		if runtime.GOARCH == "amd64" {
			archs = []string{"nt64", "nt"}
		} else {
			archs = []string{"nt"}
		}
	}
	return archs
}

func binaryPathForArch(baseboxDir, arch string) (string, bool) {
	relPath, ok := baseboxBinaryPaths[arch]
	if !ok {
		return "", false
	}

	fullPath := filepath.Join(baseboxDir, relPath)
	if exists(fullPath) {
		return relPath, true
	}
	return "", false
}
