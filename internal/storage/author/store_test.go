package author

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testAuthorID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"

func TestSaveAvatarWritesContentAddressedAuthorAsset(t *testing.T) {
	root := t.TempDir()
	data := []byte("fictional avatar bytes")
	store := New(root)

	saved, err := store.SaveAvatar(testAuthorID, "https://example.test/avatar", "image/png", data)
	if err != nil {
		t.Fatalf("SaveAvatar() error = %v", err)
	}
	wantPath := filepath.ToSlash(filepath.Join("authors", testAuthorID, "avatars", checksum(data)+".png"))
	if saved.Path != wantPath || saved.SHA256 != checksum(data) || saved.Size != int64(len(data)) {
		t.Fatalf("SaveAvatar() = %#v", saved)
	}
	stored, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(saved.Path)))
	if err != nil {
		t.Fatalf("ReadFile(avatar) error = %v", err)
	}
	if !bytes.Equal(stored, data) {
		t.Fatal("stored avatar differs from imported bytes")
	}

	again, err := store.SaveAvatar(testAuthorID, "https://other.example/avatar", "image/png", data)
	if err != nil {
		t.Fatalf("second SaveAvatar() error = %v", err)
	}
	if again.Path != saved.Path {
		t.Fatalf("duplicate bytes stored at %q and %q", saved.Path, again.Path)
	}
}

func TestSaveAvatarRejectsInvalidInput(t *testing.T) {
	store := New(t.TempDir())
	if _, err := store.SaveAvatar("not-a-uuid", "", "image/png", []byte("data")); err == nil || !strings.Contains(err.Error(), "invalid author UUID") {
		t.Fatalf("invalid UUID error = %v", err)
	}
	if _, err := store.SaveAvatar(testAuthorID, "", "text/plain", []byte("data")); err == nil || !strings.Contains(err.Error(), "invalid image content type") {
		t.Fatalf("invalid content type error = %v", err)
	}
	if _, err := store.SaveAvatar(testAuthorID, "", "image/png", nil); err == nil || !strings.Contains(err.Error(), "image data is empty") {
		t.Fatalf("empty data error = %v", err)
	}
}

func TestOpenAvatarRequiresAnAuthorScopedStoredPath(t *testing.T) {
	root := t.TempDir()
	store := New(root)
	saved, err := store.SaveAvatar(testAuthorID, "https://example.com/avatar.png", "image/png", []byte("avatar"))
	if err != nil {
		t.Fatal(err)
	}
	opened, err := store.OpenAvatar(testAuthorID, saved.Path)
	if err != nil {
		t.Fatal(err)
	}
	defer opened.File.Close()
	if opened.ContentType != "image/png" || opened.Size != int64(len("avatar")) {
		t.Fatalf("opened avatar = %#v", opened)
	}
	if _, err := store.OpenAvatar(testAuthorID, "authors/another/avatars/avatar.png"); !errors.Is(err, ErrAvatarNotFound) {
		t.Fatalf("cross-author error = %v, want ErrAvatarNotFound", err)
	}
}
