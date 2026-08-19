package document

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rudyon/malum/internal/identifier"
)

type OpenedResource struct {
	File        *os.File
	ContentType string
	Size        int64
}

func (s *Store) Load(documentID string) (Manifest, error) {
	if strings.TrimSpace(s.root) == "" {
		return Manifest{}, errors.New("document store requires a data root")
	}
	if !identifier.IsUUID(documentID) {
		return Manifest{}, fmt.Errorf("load document: invalid UUID %q", documentID)
	}
	path := filepath.Join(s.root, documentsDirectory, documentID, manifestFilename)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Manifest{}, fmt.Errorf("%w: %s", ErrDocumentNotFound, documentID)
	}
	if err != nil {
		return Manifest{}, fmt.Errorf("read document manifest: %w", err)
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode document manifest: %w", err)
	}
	if manifest.SchemaVersion != SchemaVersion {
		return Manifest{}, fmt.Errorf("load document: unsupported schema version %d", manifest.SchemaVersion)
	}
	if manifest.DocumentID != documentID {
		return Manifest{}, errors.New("load document: manifest ID does not match its bundle")
	}
	return manifest, nil
}

func (s *Store) OpenResource(documentID, filename string) (OpenedResource, error) {
	if filename == "" || filename != filepath.Base(filename) || strings.ContainsAny(filename, `/\\`) {
		return OpenedResource{}, fmt.Errorf("%w: %s", ErrResourceNotFound, filename)
	}
	manifest, err := s.Load(documentID)
	if err != nil {
		return OpenedResource{}, err
	}
	relativePath := filepath.ToSlash(filepath.Join(resourcesDirectory, filename))
	var matched *Resource
	for index := range manifest.Resources {
		resource := &manifest.Resources[index]
		if resource.Status == ResourceStored && resource.Path == relativePath {
			matched = resource
			break
		}
	}
	if matched == nil {
		return OpenedResource{}, fmt.Errorf("%w: %s", ErrResourceNotFound, filename)
	}
	path := filepath.Join(s.root, documentsDirectory, documentID, filepath.FromSlash(relativePath))
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return OpenedResource{}, fmt.Errorf("%w: %s", ErrResourceNotFound, filename)
	}
	if err != nil {
		return OpenedResource{}, fmt.Errorf("open document resource: %w", err)
	}
	return OpenedResource{File: file, ContentType: matched.ContentType, Size: matched.Size}, nil
}
