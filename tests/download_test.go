package tests

import (
	"os"
	"path/filepath"
	"testing"

	filesvc "voidfs/internal/files"
)

func TestResolveDownloadReturnsFilePath(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "docs", "manual.txt")
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filePath, []byte("manual"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	svc := filesvc.NewService(root)
	resolved, info, err := svc.ResolveDownload("docs/manual.txt")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resolved != filePath {
		t.Fatalf("expected %q, got %q", filePath, resolved)
	}
	if info.IsDir() {
		t.Fatal("expected file, got directory")
	}
}

func TestResolveDownloadRejectsDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	svc := filesvc.NewService(root)
	_, _, err := svc.ResolveDownload("docs")
	if err == nil {
		t.Fatal("expected directory error, got nil")
	}
}

func TestResolveDownloadRejectsTraversal(t *testing.T) {
	svc := filesvc.NewService(t.TempDir())
	_, _, err := svc.ResolveDownload("../../etc/passwd")
	if err == nil {
		t.Fatal("expected traversal error, got nil")
	}
}
