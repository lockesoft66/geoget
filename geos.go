package main

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

func resolveGeosArchiveRoot(baseDir string) (string, error) {
	defaultRoot := filepath.Join(baseDir, geosArchiveRoot)
	if exists(defaultRoot) {
		return defaultRoot, nil
	}

	var found string
	err := filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		if d.Name() == "geos.ini" {
			found = filepath.Dir(path)
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("search geos archive: %w", err)
	}

	if found != "" {
		return found, nil
	}

	alt := filepath.Join(baseDir, "ensemble")
	if exists(alt) {
		return alt, nil
	}

	return "", fmt.Errorf("unable to locate Ensemble archive root inside %s", baseDir)
}
