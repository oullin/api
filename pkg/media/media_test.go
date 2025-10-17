package media

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	old, _ := os.Getwd()
	os.Chdir(dir)
	t.Cleanup(func() { os.Chdir(old) })
	os.MkdirAll(GetUsersImagesDir(), 0755)

	return dir
}

func TestNewMediaAndUpload(t *testing.T) {
	setupTempDir(t)
	data := []byte{1, 2, 3}

	m, err := NewMedia("uid", data, "pic.jpg")

	if err != nil {
		t.Fatalf("new: %v", err)
	}

	if !strings.HasPrefix(m.GetFileName(), "uid-") {
		t.Fatalf("name prefix")
	}

	if m.GetExtension() != ".jpg" {
		t.Fatalf("ext")
	}

	if m.GetHeaderName() != "pic.jpg" {
		t.Fatalf("header")
	}

	if err := m.Upload(GetUsersImagesDir()); err != nil {
		t.Fatalf("upload: %v", err)
	}

	if _, err := os.Stat(m.path); err != nil {
		t.Fatalf("file not created")
	}

	if err := m.RemovePrefixedFiles(GetUsersImagesDir(), "uid"); err != nil {
		t.Fatalf("remove: %v", err)
	}
}

func TestNewMediaErrors(t *testing.T) {
	setupTempDir(t)

	if _, err := NewMedia("u", []byte{}, "a.jpg"); err == nil {
		t.Fatalf("expected empty file error")
	}

	big := make([]byte, maxFileSize+1)

	if _, err := NewMedia("u", big, "a.jpg"); err == nil {
		t.Fatalf("expected size error")
	}

	if _, err := NewMedia("u", []byte{1}, "a.txt"); err == nil {
		t.Fatalf("expected ext error")
	}
}

func TestGetFilePath(t *testing.T) {
	setupTempDir(t)
	m, err := NewMedia("u", []byte{1}, "a.jpg")

	if err != nil {
		t.Fatalf("new: %v", err)
	}

	p := m.GetFilePath("thumb")

	if !strings.Contains(filepath.Base(p), "thumb-") {
		t.Fatalf("file path wrong: %s", p)
	}
}

func TestGetPostsImagesDir(t *testing.T) {
	setupTempDir(t)

	if !strings.Contains(GetPostsImagesDir(), "posts") {
		t.Fatalf("dir invalid")
	}
}

func TestGetStorageDir(t *testing.T) {
	setupTempDir(t)

	p := GetStorageDir()

	// Get the current working directory
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	expected := filepath.Join(dir, StorageDir)
	if p != expected {
		t.Fatalf("unexpected storage dir, got: %s, want: %s", p, expected)
	}
}
