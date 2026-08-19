package author

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rudyon/malum/internal/identifier"
)

var ErrAvatarNotFound = errors.New("stored author avatar not found")

type OpenedAvatar struct {
	File        *os.File
	ContentType string
	Size        int64
}

func (s *Store) OpenAvatar(authorID, relativePath string) (OpenedAvatar, error) {
	if !identifier.IsUUID(authorID) {
		return OpenedAvatar{}, fmt.Errorf("open author avatar: invalid author UUID %q", authorID)
	}
	cleanPath := filepath.ToSlash(filepath.Clean(relativePath))
	expectedPrefix := filepath.ToSlash(filepath.Join("authors", authorID, "avatars")) + "/"
	if filepath.IsAbs(relativePath) || !strings.HasPrefix(cleanPath, expectedPrefix) || cleanPath == expectedPrefix {
		return OpenedAvatar{}, fmt.Errorf("%w: %s", ErrAvatarNotFound, relativePath)
	}
	absolutePath := filepath.Join(s.root, filepath.FromSlash(cleanPath))
	file, err := os.Open(absolutePath)
	if errors.Is(err, os.ErrNotExist) {
		return OpenedAvatar{}, fmt.Errorf("%w: %s", ErrAvatarNotFound, relativePath)
	}
	if err != nil {
		return OpenedAvatar{}, fmt.Errorf("open author avatar: %w", err)
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return OpenedAvatar{}, fmt.Errorf("inspect author avatar: %w", err)
	}
	contentType := avatarContentType(filepath.Ext(cleanPath))
	if contentType == "" {
		_ = file.Close()
		return OpenedAvatar{}, fmt.Errorf("open author avatar: unsupported stored image type %q", filepath.Ext(cleanPath))
	}
	return OpenedAvatar{File: file, ContentType: contentType, Size: info.Size()}, nil
}

func avatarContentType(extension string) string {
	switch strings.ToLower(extension) {
	case ".avif":
		return "image/avif"
	case ".gif":
		return "image/gif"
	case ".jpg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"
	default:
		return ""
	}
}
