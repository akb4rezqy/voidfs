package upload

import (
	"errors"
	"os"
	"path/filepath"

	"voidfs/internal/utils"
)

var ErrFileTooLarge = errors.New("upload exceeds max size")

type Service struct {
	root           string
	maxUploadBytes int64
}

func NewService(root string, maxUploadBytes int64) *Service {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		absRoot = root
	}
	return &Service{root: absRoot, maxUploadBytes: maxUploadBytes}
}

func (s *Service) Save(requestedPath string, content []byte) error {
	if int64(len(content)) > s.maxUploadBytes {
		return ErrFileTooLarge
	}
	resolved, err := utils.SafeJoin(s.root, requestedPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(resolved), 0o755); err != nil {
		return err
	}
	return os.WriteFile(resolved, content, 0o644)
}
