// Package author stores durable assets that belong to authors rather than documents.
package author

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rudyon/malum/internal/identifier"
)

type Store struct {
	root string
}

type SavedAvatar struct {
	SourceURL   string
	Path        string
	ContentType string
	SHA256      string
	Size        int64
}

func New(root string) *Store {
	return &Store{root: root}
}

func (s *Store) SaveAvatar(authorID, sourceURL, contentType string, data []byte) (SavedAvatar, error) {
	if strings.TrimSpace(s.root) == "" {
		return SavedAvatar{}, errors.New("author store requires a data root")
	}
	if !identifier.IsUUID(authorID) {
		return SavedAvatar{}, fmt.Errorf("store author avatar: invalid author UUID %q", authorID)
	}
	if len(data) == 0 {
		return SavedAvatar{}, errors.New("store author avatar: image data is empty")
	}
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(contentType)), "image/") {
		return SavedAvatar{}, fmt.Errorf("store author avatar: invalid image content type %q", contentType)
	}

	hash := checksum(data)
	relativePath := filepath.ToSlash(filepath.Join("authors", authorID, "avatars", hash+imageExtension(contentType)))
	absolutePath := filepath.Join(s.root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		return SavedAvatar{}, fmt.Errorf("create author avatar directory: %w", err)
	}
	if existing, err := os.ReadFile(absolutePath); err == nil {
		if checksum(existing) != hash {
			return SavedAvatar{}, fmt.Errorf("store author avatar: existing content-hash file %s is corrupt", relativePath)
		}
		return savedAvatar(sourceURL, relativePath, contentType, hash, data), nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return SavedAvatar{}, fmt.Errorf("check author avatar: %w", err)
	}

	temporary, err := os.CreateTemp(filepath.Dir(absolutePath), ".staging-")
	if err != nil {
		return SavedAvatar{}, fmt.Errorf("create author avatar staging file: %w", err)
	}
	temporaryPath := temporary.Name()
	keep := false
	defer func() {
		if !keep {
			_ = os.Remove(temporaryPath)
		}
	}()
	if err := temporary.Chmod(0o644); err != nil {
		_ = temporary.Close()
		return SavedAvatar{}, fmt.Errorf("set author avatar permissions: %w", err)
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return SavedAvatar{}, fmt.Errorf("write author avatar: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return SavedAvatar{}, fmt.Errorf("close author avatar: %w", err)
	}
	if err := os.Rename(temporaryPath, absolutePath); err != nil {
		if existing, readErr := os.ReadFile(absolutePath); readErr == nil {
			if checksum(existing) != hash {
				return SavedAvatar{}, fmt.Errorf("store author avatar: concurrently written content-hash file %s is corrupt", relativePath)
			}
			return savedAvatar(sourceURL, relativePath, contentType, hash, data), nil
		}
		return SavedAvatar{}, fmt.Errorf("publish author avatar: %w", err)
	}
	keep = true
	return savedAvatar(sourceURL, relativePath, contentType, hash, data), nil
}

func savedAvatar(sourceURL, path, contentType, hash string, data []byte) SavedAvatar {
	return SavedAvatar{
		SourceURL:   sourceURL,
		Path:        path,
		ContentType: contentType,
		SHA256:      hash,
		Size:        int64(len(data)),
	}
}

func checksum(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func imageExtension(contentType string) string {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/avif":
		return ".avif"
	case "image/gif":
		return ".gif"
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/svg+xml":
		return ".svg"
	case "image/webp":
		return ".webp"
	default:
		return ".img"
	}
}
