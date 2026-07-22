package editor

import (
	"errors"
	"os"
	"path/filepath"

	"voidfs/internal/utils"
)

var ErrFileTooLarge = errors.New("file exceeds editor size limit")

type Service struct {
	root         string
	maxEditBytes int64
}

func NewService(root string, maxEditBytes int64) *Service {
	return &Service{root: root, maxEditBytes: maxEditBytes}
}

func (s *Service) Open(requestedPath string) (Document, error) {
	resolved, err := utils.SafeJoin(s.root, requestedPath)
	if err != nil {
		return Document{}, err
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return Document{}, err
	}
	if info.Size() > s.maxEditBytes {
		return Document{}, ErrFileTooLarge
	}

	content, err := os.ReadFile(resolved)
	if err != nil {
		return Document{}, err
	}

	return Document{
		Path:    filepath.ToSlash(requestedPath),
		Content: string(content),
		Dirty:   false,
	}, nil
}

func (s *Service) Save(requestedPath string, content string) error {
	resolved, err := utils.SafeJoin(s.root, requestedPath)
	if err != nil {
		return err
	}
	if int64(len(content)) > s.maxEditBytes {
		return ErrFileTooLarge
	}
	if err := os.MkdirAll(filepath.Dir(resolved), 0o755); err != nil {
		return err
	}
	return writeFileAtomic(resolved, []byte(content), 0o644)
}
