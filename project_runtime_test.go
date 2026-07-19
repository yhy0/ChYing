package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestValidateExistingProject(t *testing.T) {
	root := filepath.Join(t.TempDir(), "db")
	projectDir := filepath.Join(root, "src-auto")
	if err := os.MkdirAll(projectDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "src-auto.db"), []byte("sqlite"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := validateExistingProjectAt(root, "src-auto"); err != nil {
		t.Fatalf("valid project rejected: %v", err)
	}
}

func TestValidateExistingProjectRejectsSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires additional privileges on Windows")
	}
	root := filepath.Join(t.TempDir(), "db")
	target := filepath.Join(t.TempDir(), "project")
	if err := os.MkdirAll(target, 0o700); err != nil {
		t.Fatal(err)
	}
	projectDir := filepath.Join(root, "linked")
	if err := os.MkdirAll(filepath.Dir(projectDir), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, projectDir); err != nil {
		t.Fatal(err)
	}
	if err := validateExistingProjectAt(root, "linked"); err == nil {
		t.Fatal("expected symlink project to be rejected")
	}
}

func TestResultErrorRejectsMissingStepResult(t *testing.T) {
	if err := resultError(Result{}); err == nil {
		t.Fatal("expected missing initialization result to fail closed")
	}
}
