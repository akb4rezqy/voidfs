package tests

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	filesvc "voidfs/internal/files"
)

func TestListDirectoryReturnsSortedEntries(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "b-folder"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	svc := filesvc.NewService(root)
	entries, err := svc.List("/")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if !entries[0].IsDir || entries[0].Name != "b-folder" {
		t.Fatalf("expected directory first, got %+v", entries[0])
	}
	if entries[1].IsDir || entries[1].Name != "a.txt" {
		t.Fatalf("expected file second, got %+v", entries[1])
	}
}

func TestListRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	svc := filesvc.NewService(root)

	_, err := svc.List("../../etc")
	if err == nil {
		t.Fatal("expected traversal error, got nil")
	}
}

func TestCreateFolderMakesDirectory(t *testing.T) {
	root := t.TempDir()
	svc := filesvc.NewService(root)

	if err := svc.CreateFolder("projects/new-app"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	info, err := os.Stat(filepath.Join(root, "projects", "new-app"))
	if err != nil {
		t.Fatalf("stat created dir: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected created path to be directory")
	}
}

func TestCreateFileMakesEmptyFile(t *testing.T) {
	root := t.TempDir()
	svc := filesvc.NewService(root)

	if err := svc.CreateFile("projects/main.go"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	info, err := os.Stat(filepath.Join(root, "projects", "main.go"))
	if err != nil {
		t.Fatalf("stat created file: %v", err)
	}
	if info.IsDir() || info.Size() != 0 {
		t.Fatalf("expected empty file, got mode %v and size %d", info.Mode(), info.Size())
	}
}

func TestCreateFileDoesNotOverwriteExistingFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "notes.txt")
	if err := os.WriteFile(path, []byte("keep me"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	svc := filesvc.NewService(root)
	if err := svc.CreateFile("notes.txt"); err == nil {
		t.Fatal("expected existing file error, got nil")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read existing file: %v", err)
	}
	if string(content) != "keep me" {
		t.Fatalf("existing content was changed: %q", content)
	}
}

func TestRenameMovesFile(t *testing.T) {
	root := t.TempDir()
	oldPath := filepath.Join(root, "old.txt")
	if err := os.WriteFile(oldPath, []byte("halo"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	svc := filesvc.NewService(root)
	if err := svc.Rename("old.txt", "renamed.txt"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "renamed.txt")); err != nil {
		t.Fatalf("renamed file missing: %v", err)
	}
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("expected old file gone, got %v", err)
	}
}

func TestDeleteRemovesPath(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "trash.txt")
	if err := os.WriteFile(path, []byte("bye"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	svc := filesvc.NewService(root)
	if err := svc.Delete("trash.txt"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected deleted file gone, got %v", err)
	}
}

func TestDeleteManyRemovesSelectedPaths(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"one.txt", "two.txt", "keep.txt"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(name), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	svc := filesvc.NewService(root)
	if err := svc.DeleteMany([]string{"one.txt", "two.txt"}); err != nil {
		t.Fatalf("delete many: %v", err)
	}
	for _, name := range []string{"one.txt", "two.txt"} {
		if _, err := os.Stat(filepath.Join(root, name)); !os.IsNotExist(err) {
			t.Fatalf("expected %s removed, got %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "keep.txt")); err != nil {
		t.Fatalf("expected keep.txt to remain: %v", err)
	}
}

func TestCreateZipAndExtractSelectedPaths(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "project", "nested"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "project", "nested", "main.go"), []byte("package main"), 0o644); err != nil {
		t.Fatalf("write nested file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "notes.txt"), []byte("notes"), 0o644); err != nil {
		t.Fatalf("write notes: %v", err)
	}

	svc := filesvc.NewService(root)
	if err := svc.CreateZip([]string{"project", "notes.txt"}, "backup.zip"); err != nil {
		t.Fatalf("create zip: %v", err)
	}
	reader, err := zip.OpenReader(filepath.Join(root, "backup.zip"))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	names := make(map[string]bool)
	for _, entry := range reader.File {
		names[entry.Name] = true
	}
	reader.Close()
	for _, name := range []string{"project/", "project/nested/", "project/nested/main.go", "notes.txt"} {
		if !names[name] {
			t.Fatalf("archive entry missing: %s", name)
		}
	}

	if err := os.RemoveAll(filepath.Join(root, "project")); err != nil {
		t.Fatalf("remove source folder: %v", err)
	}
	if err := os.Remove(filepath.Join(root, "notes.txt")); err != nil {
		t.Fatalf("remove source file: %v", err)
	}
	if err := svc.ExtractZips([]string{"backup.zip"}); err != nil {
		t.Fatalf("extract zip: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(root, "project", "nested", "main.go"))
	if err != nil {
		t.Fatalf("read extracted file: %v", err)
	}
	if string(content) != "package main" {
		t.Fatalf("unexpected extracted content: %q", content)
	}
}

func TestExtractZipRejectsExistingPath(t *testing.T) {
	root := t.TempDir()
	archivePath := filepath.Join(root, "files.zip")
	createTestZip(t, archivePath, "notes.txt", "from archive")
	if err := os.WriteFile(filepath.Join(root, "notes.txt"), []byte("keep me"), 0o644); err != nil {
		t.Fatalf("write existing file: %v", err)
	}

	svc := filesvc.NewService(root)
	if err := svc.ExtractZips([]string{"files.zip"}); err == nil {
		t.Fatal("expected existing path error, got nil")
	}
	content, err := os.ReadFile(filepath.Join(root, "notes.txt"))
	if err != nil {
		t.Fatalf("read existing file: %v", err)
	}
	if string(content) != "keep me" {
		t.Fatalf("existing file was overwritten: %q", content)
	}
}

func TestExtractZipRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	archivePath := filepath.Join(root, "unsafe.zip")
	createTestZip(t, archivePath, "../outside.txt", "unsafe")

	svc := filesvc.NewService(root)
	if err := svc.ExtractZips([]string{"unsafe.zip"}); err == nil {
		t.Fatal("expected traversal error, got nil")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(root), "outside.txt")); !os.IsNotExist(err) {
		t.Fatalf("archive escaped root: %v", err)
	}
}

func TestExtractZipRejectsSymlinkParent(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "linked")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	archivePath := filepath.Join(root, "unsafe.zip")
	createTestZip(t, archivePath, "linked/outside.txt", "unsafe")

	svc := filesvc.NewService(root)
	if err := svc.ExtractZips([]string{"unsafe.zip"}); err == nil {
		t.Fatal("expected symlink parent error, got nil")
	}
	if _, err := os.Stat(filepath.Join(outside, "outside.txt")); !os.IsNotExist(err) {
		t.Fatalf("archive escaped through symlink: %v", err)
	}
}

func createTestZip(t *testing.T, path, name, content string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	writer := zip.NewWriter(file)
	entry, err := writer.Create(name)
	if err != nil {
		t.Fatalf("create zip entry: %v", err)
	}
	if _, err := entry.Write([]byte(content)); err != nil {
		t.Fatalf("write zip entry: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close zip file: %v", err)
	}
}
