package tests

import (
	"os"
	"testing"

	"voidfs/internal/config"
)

func TestLoadUsesDefaults(t *testing.T) {
	for _, key := range []string{
		"APP_ADDR",
		"APP_ROOT_DIR",
		"APP_SESSION_SECRET",
		"APP_MAX_UPLOAD_BYTES",
		"APP_MAX_EDIT_BYTES",
		"APP_ALLOWED_USER",
	} {
		t.Setenv(key, "")
	}

	cfg := config.Load()
	if cfg.Addr != ":8787" {
		t.Fatalf("expected default addr, got %q", cfg.Addr)
	}
	if cfg.RootDir != "/" {
		t.Fatalf("expected filesystem root, got %q", cfg.RootDir)
	}
	if cfg.AllowedUser != "root" {
		t.Fatalf("expected root-only login, got %q", cfg.AllowedUser)
	}
	if cfg.MaxUploadBytes != 10*1024*1024 {
		t.Fatalf("expected default upload bytes, got %d", cfg.MaxUploadBytes)
	}
}

func TestLoadUsesEnvOverrides(t *testing.T) {
	_ = os.Setenv("APP_ADDR", ":9090")
	_ = os.Setenv("APP_ROOT_DIR", "/srv/files")
	_ = os.Setenv("APP_MAX_UPLOAD_BYTES", "2048")
	t.Cleanup(func() {
		_ = os.Unsetenv("APP_ADDR")
		_ = os.Unsetenv("APP_ROOT_DIR")
		_ = os.Unsetenv("APP_MAX_UPLOAD_BYTES")
	})

	cfg := config.Load()
	if cfg.Addr != ":9090" {
		t.Fatalf("expected env addr, got %q", cfg.Addr)
	}
	if cfg.RootDir != "/srv/files" {
		t.Fatalf("expected env root, got %q", cfg.RootDir)
	}
	if cfg.MaxUploadBytes != 2048 {
		t.Fatalf("expected env upload bytes, got %d", cfg.MaxUploadBytes)
	}
}
