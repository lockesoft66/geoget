package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func createLaunchers(installRoot string) error {
	templDir, err := templateDir()
	if err != nil {
		return err
	}

	launchers := []struct {
		templateName string
		outputName   string
		executable   bool
	}{
		{"ensemble.sh", "ensemble.sh", true},
		{"ensemble.cmd", "ensemble.cmd", false},
		{"ensemble.ps1", "ensemble.ps1", false},
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
	executables := []string{
		filepath.Join(baseboxDir, "binl64", "basebox"),
		filepath.Join(baseboxDir, "binl", "basebox"),
		filepath.Join(baseboxDir, "binmac", "basebox"),
		filepath.Join(baseboxDir, "binnt", "basebox.exe"),
	}

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		for _, exe := range executables {
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

func detectBaseboxBinary(baseboxDir string) string {
	switch runtime.GOOS {
	case "linux":
		if exists(filepath.Join(baseboxDir, "binl64", "basebox")) {
			return filepath.Join("binl64", "basebox")
		}
		if exists(filepath.Join(baseboxDir, "binl", "basebox")) {
			return filepath.Join("binl", "basebox")
		}
	case "darwin":
		if exists(filepath.Join(baseboxDir, "binmac", "basebox")) {
			return filepath.Join("binmac", "basebox")
		}
	case "windows":
		if exists(filepath.Join(baseboxDir, "binnt", "basebox.exe")) {
			return filepath.Join("binnt", "basebox.exe")
		}
	}

	// Fallback ordering if GOOS detection did not match available binaries.
	fallback := []string{
		filepath.Join("binl64", "basebox"),
		filepath.Join("binl", "basebox"),
		filepath.Join("binmac", "basebox"),
		filepath.Join("binnt", "basebox.exe"),
	}

	for _, candidate := range fallback {
		if exists(filepath.Join(baseboxDir, candidate)) {
			return candidate
		}
	}

	return ""
}

func templateDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate executable: %w", err)
	}
	return filepath.Join(filepath.Dir(exe), "templ"), nil
}
