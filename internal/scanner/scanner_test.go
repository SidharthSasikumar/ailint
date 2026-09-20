package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMissingRootFails(t *testing.T) {
	if _, err := New(filepath.Join(t.TempDir(), "missing"), nil).Scan(); err == nil {
		t.Fatal("missing root reported success")
	}
}
func TestSkipsExternalSymlinks(t *testing.T) {
	outside := filepath.Join(t.TempDir(), "secret.go")
	os.WriteFile(outside, []byte("package secret"), 0600)
	root := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "linked.go")); err != nil {
		t.Fatal(err)
	}
	files, err := New(root, nil).Scan()
	if err != nil || len(files) != 0 {
		t.Fatalf("external file scanned: %v %v", files, err)
	}
}
