package tests

import (
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
