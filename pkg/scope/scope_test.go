package scope

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidator_Safe_Inside(t *testing.T) {
	dir := t.TempDir()
	v, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	sub := filepath.Join(dir, "sub", "file.txt")
	resolved, err := v.Safe(sub)
	if err != nil {
		t.Fatalf("esperaba OK, got: %v", err)
	}
	if resolved != filepath.Clean(sub) {
		t.Fatalf("esperaba %s, got %s", sub, resolved)
	}
}

func TestValidator_Safe_Outside(t *testing.T) {
	dir := t.TempDir()
	v, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	_, err = v.Safe(os.TempDir())
	if err == nil {
		t.Fatal("esperaba error para path fuera del proyecto")
	}
}

func TestValidator_Safe_PathTraversal(t *testing.T) {
	dir := t.TempDir()
	v, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	_, err = v.Safe(filepath.Join(dir, "..", "..", "etc", "passwd"))
	if err == nil {
		t.Fatal("esperaba error para path traversal")
	}
}

func TestValidator_New_NonExistent(t *testing.T) {
	_, err := New("/ruta/que/no/existe/xyz123")
	if err == nil {
		t.Fatal("esperaba error para directorio inexistente")
	}
}
