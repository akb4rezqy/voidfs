package tests

import (
	"path/filepath"
	"testing"

	"voidfs/internal/utils"
)

func TestSafeJoinAllowsRoot(t *testing.T) {
	root := t.TempDir()

	got, err := utils.SafeJoin(root, "/")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	expected, _ := filepath.Abs(root)
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestSafeJoinAllowsNestedPath(t *testing.T) {
	root := t.TempDir()

	got, err := utils.SafeJoin(root, "folder/file.txt")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	expected := filepath.Join(root, "folder", "file.txt")
	expected, _ = filepath.Abs(expected)
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestSafeJoinAllowsNestedPathWhenRootIsFilesystemRoot(t *testing.T) {
	got, err := utils.SafeJoin("/", "root/portfolio")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != "/root/portfolio" {
		t.Fatalf("expected /root/portfolio, got %q", got)
	}
}

func TestSafeJoinRejectsTraversal(t *testing.T) {
	root := t.TempDir()

	_, err := utils.SafeJoin(root, "../../etc/passwd")
	if err == nil {
		t.Fatal("expected traversal error, got nil")
	}

	if err != utils.ErrPathOutsideRoot {
		t.Fatalf("expected ErrPathOutsideRoot, got %v", err)
	}
}
