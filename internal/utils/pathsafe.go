package utils

import (
	"errors"
	"path/filepath"
	"strings"
)

var ErrPathOutsideRoot = errors.New("path escapes root directory")

func SafeJoin(root, requested string) (string, error) {
	cleanRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}

	if requested == "" || requested == "." || requested == "/" {
		return cleanRoot, nil
	}

	normalized := strings.TrimPrefix(requested, string(filepath.Separator))
	cleanRequested := filepath.Clean(normalized)
	if cleanRequested == ".." || strings.HasPrefix(cleanRequested, ".."+string(filepath.Separator)) {
		return "", ErrPathOutsideRoot
	}

	candidate := filepath.Join(cleanRoot, cleanRequested)
	candidate, err = filepath.Abs(candidate)
	if err != nil {
		return "", err
	}

	if candidate == cleanRoot {
		return candidate, nil
	}

	prefix := cleanRoot
	if cleanRoot != string(filepath.Separator) {
		prefix += string(filepath.Separator)
	}
	if !strings.HasPrefix(candidate, prefix) {
		return "", ErrPathOutsideRoot
	}

	return candidate, nil
}
