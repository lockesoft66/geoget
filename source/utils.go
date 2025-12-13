package main

import (
	"os"
	"path/filepath"
)

func filepathDir(path string) string {
	dir := filepath.Dir(path)
	if dir == "." {
		return ""
	}
	return dir
}

func resolveBaseboxRoot(baseDir string) string {
	candidate := filepath.Join(baseDir, "pcgeos-basebox")
	if exists(candidate) {
		return candidate
	}
	return baseDir
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
