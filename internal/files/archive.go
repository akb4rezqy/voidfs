package files

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"voidfs/internal/utils"
)

func (s *Service) DeleteMany(requestedPaths []string) error {
	resolvedPaths, err := s.resolvePaths(requestedPaths)
	if err != nil {
		return err
	}
	for _, resolved := range resolvedPaths {
		if err := os.RemoveAll(resolved); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) CreateZip(requestedPaths []string, destination string) (err error) {
	resolvedPaths, err := s.resolvePaths(requestedPaths)
	if err != nil {
		return err
	}
	resolvedDestination, err := utils.SafeJoin(s.root, destination)
	if err != nil {
		return err
	}
	if resolvedDestination == s.root {
		return errors.New("archive destination is required")
	}
	for _, resolved := range resolvedPaths {
		info, statErr := os.Lstat(resolved)
		if statErr != nil {
			return statErr
		}
		if info.IsDir() && isWithin(resolved, resolvedDestination) {
			return errors.New("archive cannot be created inside a selected folder")
		}
	}

	archiveFile, err := os.OpenFile(resolvedDestination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := archiveFile.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			_ = os.Remove(resolvedDestination)
		}
	}()

	writer := zip.NewWriter(archiveFile)
	defer func() {
		if closeErr := writer.Close(); err == nil {
			err = closeErr
		}
	}()

	seen := make(map[string]bool)
	for _, resolved := range resolvedPaths {
		base := filepath.Dir(resolved)
		walkErr := filepath.Walk(resolved, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("symbolic links are not supported: %s", info.Name())
			}
			name, relErr := filepath.Rel(base, path)
			if relErr != nil {
				return relErr
			}
			name = filepath.ToSlash(name)
			if info.IsDir() {
				name += "/"
			}
			if seen[name] {
				return fmt.Errorf("duplicate archive entry: %s", name)
			}
			seen[name] = true

			header, headerErr := zip.FileInfoHeader(info)
			if headerErr != nil {
				return headerErr
			}
			header.Name = name
			header.Method = zip.Deflate
			entry, createErr := writer.CreateHeader(header)
			if createErr != nil || info.IsDir() {
				return createErr
			}
			file, openErr := os.Open(path)
			if openErr != nil {
				return openErr
			}
			_, copyErr := io.Copy(entry, file)
			closeErr := file.Close()
			if copyErr != nil {
				return copyErr
			}
			return closeErr
		})
		if walkErr != nil {
			return walkErr
		}
	}
	return nil
}

func (s *Service) ExtractZips(requestedPaths []string) error {
	resolvedPaths, err := s.resolvePaths(requestedPaths)
	if err != nil {
		return err
	}
	for _, resolved := range resolvedPaths {
		if !strings.EqualFold(filepath.Ext(resolved), ".zip") {
			return fmt.Errorf("not a zip archive: %s", filepath.Base(resolved))
		}
		if err := extractZip(resolved, filepath.Dir(resolved)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) resolvePaths(requestedPaths []string) ([]string, error) {
	if len(requestedPaths) == 0 {
		return nil, errors.New("at least one path is required")
	}
	resolvedPaths := make([]string, 0, len(requestedPaths))
	seen := make(map[string]bool)
	for _, requested := range requestedPaths {
		resolved, err := utils.SafeJoin(s.root, requested)
		if err != nil {
			return nil, err
		}
		if resolved == s.root {
			return nil, errors.New("root directory cannot be selected")
		}
		if !seen[resolved] {
			seen[resolved] = true
			resolvedPaths = append(resolvedPaths, resolved)
		}
	}
	return resolvedPaths, nil
}

func extractZip(archivePath, destination string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	targets := make([]string, len(reader.File))
	for i, entry := range reader.File {
		if entry.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symbolic links are not supported: %s", entry.Name)
		}
		target, err := safeArchiveTarget(destination, entry.Name)
		if err != nil {
			return err
		}
		if err := rejectSymlinkParents(destination, target); err != nil {
			return err
		}
		if _, err := os.Lstat(target); err == nil {
			return fmt.Errorf("path already exists: %s", entry.Name)
		} else if !os.IsNotExist(err) {
			return err
		}
		targets[i] = target
	}

	for i, entry := range reader.File {
		target := targets[i]
		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(target, entry.Mode().Perm()); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		source, err := entry.Open()
		if err != nil {
			return err
		}
		destinationFile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, entry.Mode().Perm())
		if err != nil {
			source.Close()
			return err
		}
		_, copyErr := io.Copy(destinationFile, source)
		sourceCloseErr := source.Close()
		destinationCloseErr := destinationFile.Close()
		if copyErr != nil {
			return copyErr
		}
		if sourceCloseErr != nil {
			return sourceCloseErr
		}
		if destinationCloseErr != nil {
			return destinationCloseErr
		}
	}
	return nil
}

func safeArchiveTarget(destination, name string) (string, error) {
	cleanName := filepath.Clean(filepath.FromSlash(name))
	if cleanName == "." || filepath.IsAbs(cleanName) || cleanName == ".." || strings.HasPrefix(cleanName, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid archive path: %s", name)
	}
	target := filepath.Join(destination, cleanName)
	if !isWithin(destination, target) {
		return "", fmt.Errorf("archive path escapes destination: %s", name)
	}
	return target, nil
}

func isWithin(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func rejectSymlinkParents(destination, target string) error {
	rel, err := filepath.Rel(destination, filepath.Dir(target))
	if err != nil {
		return err
	}
	current := destination
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		if part == "." || part == "" {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("archive path uses symbolic link: %s", part)
		}
	}
	return nil
}
