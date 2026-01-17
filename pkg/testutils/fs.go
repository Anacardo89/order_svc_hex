package testutils

import (
	"os"
	"path/filepath"
)

func FindDevRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", os.ErrNotExist
}

func BuildPath(location string) (string, error) {
	root, err := FindDevRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, location), nil
}
