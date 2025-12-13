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
	baseboxArchFallbackOrder = []string{"l64", "rpi64", "mac", "nt64", "nt"}
)

func createLaunchers(installRoot, arch string) error {
	templDir, err := templateDir()
	if err != nil {
		return err
	}

	launchers, err := launcherTemplatesForArch(arch)
	if err != nil {
		return err
	}

	for _, launcher := range launchers {
		source := filepath.Join(templDir, launcher.templateName)
		dest := filepath.Join(installRoot, launcher.outputName)

		content, err := os.ReadFile(source)
		if err != nil {
			return fmt.Errorf("read launcher template %s: %w", source, err)
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
	templDir, err := templateDir()
	if err != nil {
		return err
	}

	templatePath := filepath.Join(templDir, "basebox.conf")
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read basebox template: %w", err)
	}

	hostPath := filepath.Clean(drivecDir)
	config := strings.ReplaceAll(string(data), "{{TAG}}", hostPath)

	dest := filepath.Join(baseboxDir, "basebox.conf")
	if err := os.WriteFile(dest, []byte(config), 0o644); err != nil {
		return fmt.Errorf("write basebox.conf: %w", err)
	}

	return nil
}

func ensureExecutables(baseboxDir string) error {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		for _, arch := range baseboxArchFallbackOrder {
			relPath, ok := baseboxBinaryPaths[arch]
			if !ok {
				continue
			}
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
	var preferred []string

	switch runtime.GOOS {
	case "linux":
		if runtime.GOARCH == "arm64" {
			preferred = append(preferred, "rpi64", "l64")
		} else {
			preferred = append(preferred, "l64", "rpi64")
		}
	case "darwin":
		preferred = append(preferred, "mac", "l64")
	case "windows":
		if runtime.GOARCH == "amd64" {
			preferred = append(preferred, "nt64")
		}
		preferred = append(preferred, "nt")
	}

	return appendMissingArch(preferred, baseboxArchFallbackOrder)
}

func appendMissingArch(prefix, suffix []string) []string {
	seen := make(map[string]struct{}, len(prefix)+len(suffix))
	result := make([]string, 0, len(prefix)+len(suffix))

	for _, arch := range prefix {
		if _, ok := seen[arch]; ok {
			continue
		}
		seen[arch] = struct{}{}
		result = append(result, arch)
	}

	for _, arch := range suffix {
		if _, ok := seen[arch]; ok {
			continue
		}
		seen[arch] = struct{}{}
		result = append(result, arch)
	}

	return result
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

func templateDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate executable: %w", err)
	}
	exeDir := filepath.Dir(exe)

	candidates := []string{
		filepath.Join(exeDir, "templ"),                 // old layout (optional fallback)
		filepath.Join(exeDir, "source", "templ"),       // source/templ under repo root
		filepath.Join(exeDir, "..", "source", "templ"), // fallback if binary ends up in a subdir
	}

	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir, nil
		}
	}

	return "", fmt.Errorf("unable to locate templ directory (checked %v)", candidates)
}
