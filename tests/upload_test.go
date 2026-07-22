package tests

import (
	"os"
	"path/filepath"
	"testing"

	uploadsvc "voidfs/internal/upload"
)

func TestUploadWritesFile(t *testing.T) {
	root := t.TempDir()
	svc := uploadsvc.NewService(root, 1024)

	if err := svc.Save("docs/readme.txt", []byte("upload ok")); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	content, err := os.ReadFile(filepath.Join(root, "docs", "readme.txt"))
	if err != nil {
		t.Fatalf("read uploaded file: %v", err)
	}
	if string(content) != "upload ok" {
		t.Fatalf("expected upload ok, got %q", string(content))
	}
}

func TestUploadRejectsLargeFile(t *testing.T) {
	root := t.TempDir()
	svc := uploadsvc.NewService(root, 3)

	err := svc.Save("big.txt", []byte("1234"))
	if err == nil {
		t.Fatal("expected size error, got nil")
	}
}
