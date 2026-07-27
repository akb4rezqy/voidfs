package files

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"voidfs/internal/utils"
)

type Service struct {
	root string
}

func NewService(root string) *Service {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		absRoot = root
	}
	return &Service{root: absRoot}
}

func (s *Service) List(requestedPath string) ([]Entry, error) {
	resolved, err := utils.SafeJoin(s.root, requestedPath)
	if err != nil {
		return nil, err
	}

	dirEntries, err := os.ReadDir(resolved)
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, 0, len(dirEntries))
	for _, item := range dirEntries {
		info, err := item.Info()
		if err != nil {
			return nil, err
		}

		relPath, err := filepath.Rel(s.root, filepath.Join(resolved, item.Name()))
		if err != nil {
			return nil, err
		}
		relPath = filepath.ToSlash(relPath)
		if relPath == "." {
			relPath = ""
		}

		entries = append(entries, Entry{
			Name:    item.Name(),
			Path:    relPath,
			Size:    info.Size(),
			IsDir:   item.IsDir(),
			ModTime: info.ModTime(),
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})

	return entries, nil
}

func (s *Service) CreateFolder(requestedPath string) error {
	resolved, err := utils.SafeJoin(s.root, requestedPath)
	if err != nil {
		return err
	}
	return os.MkdirAll(resolved, 0o755)
}

func (s *Service) CreateFile(requestedPath string) error {
	resolved, err := utils.SafeJoin(s.root, requestedPath)
	if err != nil {
		return err
	}
	if resolved == s.root {
		return errors.New("file path is required")
	}
	if err := os.MkdirAll(filepath.Dir(resolved), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(resolved, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	return file.Close()
}

func (s *Service) Rename(oldPath, newPath string) error {
	resolvedOld, err := utils.SafeJoin(s.root, oldPath)
	if err != nil {
		return err
	}
	resolvedNew, err := utils.SafeJoin(s.root, newPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(resolvedNew), 0o755); err != nil {
		return err
	}
	return os.Rename(resolvedOld, resolvedNew)
}

func (s *Service) Delete(requestedPath string) error {
	resolved, err := utils.SafeJoin(s.root, requestedPath)
	if err != nil {
		return err
	}
	return os.RemoveAll(resolved)
}

func (s *Service) ResolveDownload(requestedPath string) (string, os.FileInfo, error) {
	resolved, err := utils.SafeJoin(s.root, requestedPath)
	if err != nil {
		return "", nil, err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", nil, err
	}
	if info.IsDir() {
		return "", nil, errors.New("directories cannot be downloaded")
	}
	return resolved, info, nil
}
