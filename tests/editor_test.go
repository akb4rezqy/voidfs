package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	editorsvc "voidfs/internal/editor"
)

func TestOpenReadsTextFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "main.go")
	if err := os.WriteFile(path, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	svc := editorsvc.NewService(root, 1024)
	doc, err := svc.Open("main.go")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if doc.Path != "main.go" {
		t.Fatalf("expected path main.go, got %q", doc.Path)
	}
	if !strings.Contains(doc.Content, "package main") {
		t.Fatalf("expected file content, got %q", doc.Content)
	}
}

func TestSaveWritesFile(t *testing.T) {
	root := t.TempDir()
	svc := editorsvc.NewService(root, 1024)

	if err := svc.Save("notes.txt", "halo"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	content, err := os.ReadFile(filepath.Join(root, "notes.txt"))
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}
	if string(content) != "halo" {
		t.Fatalf("expected halo, got %q", string(content))
	}
}

func TestOpenRejectsLargeFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "big.txt")
	if err := os.WriteFile(path, []byte(strings.Repeat("a", 2048)), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	svc := editorsvc.NewService(root, 128)
	_, err := svc.Open("big.txt")
	if err == nil {
		t.Fatal("expected size error, got nil")
	}
}
