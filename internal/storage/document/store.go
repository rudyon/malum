package document

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rudyon/malum/internal/identifier"
	"github.com/rudyon/malum/internal/ingest/webpage"
)

const (
	documentsDirectory = "documents"
	manifestFilename   = "document.json"
	originalFilename   = "original.html"
	articleFilename    = "article.html"
	resourcesDirectory = "resources"
)

type Store struct {
	root  string
	now   func() time.Time
	newID func() (string, error)
}

func New(root string) *Store {
	return &Store{
		root:  root,
		now:   time.Now,
		newID: identifier.NewUUID,
	}
}

func (s *Store) SaveWebpage(result webpage.Result) (saved Saved, err error) {
	if err := validateResult(result); err != nil {
		return Saved{}, err
	}
	if strings.TrimSpace(s.root) == "" {
		return Saved{}, errors.New("document store requires a data root")
	}

	documentID, err := s.newID()
	if err != nil {
		return Saved{}, fmt.Errorf("generate document ID: %w", err)
	}
	if !identifier.IsUUID(documentID) {
		return Saved{}, fmt.Errorf("generate document ID: invalid UUID %q", documentID)
	}

	documentsPath := filepath.Join(s.root, documentsDirectory)
	if err := os.MkdirAll(documentsPath, 0o755); err != nil {
		return Saved{}, fmt.Errorf("create documents directory: %w", err)
	}
	finalPath := filepath.Join(documentsPath, documentID)
	if _, err := os.Stat(finalPath); err == nil {
		return Saved{}, fmt.Errorf("store document: ID %s already exists", documentID)
	} else if !errors.Is(err, os.ErrNotExist) {
		return Saved{}, fmt.Errorf("check document path: %w", err)
	}

	stagingPath, err := os.MkdirTemp(documentsPath, ".staging-")
	if err != nil {
		return Saved{}, fmt.Errorf("create document staging directory: %w", err)
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(stagingPath)
		}
	}()

	originalHash := checksum(result.Snapshot.OriginalHTML)
	if err = writeFile(stagingPath, originalFilename, result.Snapshot.OriginalHTML); err != nil {
		return Saved{}, err
	}

	resources, storedPaths, err := storeResources(stagingPath, result.Document.Resources)
	if err != nil {
		return Saved{}, err
	}
	articleHTML, err := rewriteResourcePaths(result.Document.ContentHTML, storedPaths)
	if err != nil {
		return Saved{}, err
	}
	articleHash := checksum(articleHTML)
	if err = writeFile(stagingPath, articleFilename, articleHTML); err != nil {
		return Saved{}, err
	}

	manifest := buildManifest(documentID, s.now().UTC(), result, originalHash, articleHTML, articleHash, resources)
	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return Saved{}, fmt.Errorf("encode document manifest: %w", err)
	}
	manifestJSON = append(manifestJSON, '\n')
	if err = writeFile(stagingPath, manifestFilename, manifestJSON); err != nil {
		return Saved{}, err
	}

	if err = os.Rename(stagingPath, finalPath); err != nil {
		return Saved{}, fmt.Errorf("publish document bundle: %w", err)
	}

	return Saved{ID: documentID, Path: finalPath, Manifest: manifest}, nil
}

func validateResult(result webpage.Result) error {
	if len(result.Snapshot.OriginalHTML) == 0 {
		return errors.New("store webpage: original HTML is empty")
	}
	if strings.TrimSpace(result.Snapshot.RequestedURL) == "" || strings.TrimSpace(result.Snapshot.FinalURL) == "" {
		return errors.New("store webpage: acquisition URLs are required")
	}
	if strings.TrimSpace(result.Document.ContentHTML) == "" {
		return errors.New("store webpage: normalized article HTML is empty")
	}
	if strings.TrimSpace(result.Document.Title) == "" {
		return errors.New("store webpage: article title is required")
	}
	return nil
}

func storeResources(stagingPath string, imported []webpage.Resource) ([]Resource, map[string]string, error) {
	resources := make([]Resource, 0, len(imported))
	storedPaths := make(map[string]string)
	pathsByHash := make(map[string]string)

	for _, importedResource := range imported {
		resource := Resource{
			SourceURL: importedResource.URL,
			Role:      importedResource.Role,
			Status:    ResourceUnavailable,
		}
		if len(importedResource.Data) == 0 {
			resources = append(resources, resource)
			continue
		}
		if !strings.HasPrefix(strings.ToLower(importedResource.ContentType), "image/") {
			return nil, nil, fmt.Errorf("store resource %q: invalid image content type %q", importedResource.URL, importedResource.ContentType)
		}

		hash := checksum(importedResource.Data)
		relativePath, exists := pathsByHash[hash]
		if !exists {
			filename := hash + imageExtension(importedResource.ContentType)
			relativePath = filepath.ToSlash(filepath.Join(resourcesDirectory, filename))
			if err := writeFile(stagingPath, relativePath, importedResource.Data); err != nil {
				return nil, nil, err
			}
			pathsByHash[hash] = relativePath
		}

		resource.Status = ResourceStored
		resource.Path = relativePath
		resource.ContentType = importedResource.ContentType
		resource.SHA256 = hash
		resource.Size = int64(len(importedResource.Data))
		resources = append(resources, resource)
		storedPaths[importedResource.URL] = relativePath
	}

	return resources, storedPaths, nil
}

func buildManifest(
	documentID string,
	storedAt time.Time,
	result webpage.Result,
	originalHash string,
	articleHTML []byte,
	articleHash string,
	resources []Resource,
) Manifest {
	document := result.Document
	return Manifest{
		SchemaVersion: SchemaVersion,
		DocumentID:    documentID,
		ReadingKind:   ReadingKindArticle,
		StoredAt:      storedAt,
		Acquisition: Acquisition{
			Method:       AcquisitionURL,
			RequestedURL: result.Snapshot.RequestedURL,
			FinalURL:     result.Snapshot.FinalURL,
		},
		Original: Original{
			Format:      OriginalFormatHTML,
			ContentType: result.Snapshot.ContentType,
			Path:        originalFilename,
			SHA256:      originalHash,
			Size:        int64(len(result.Snapshot.OriginalHTML)),
		},
		Article: Article{
			Path:               articleFilename,
			SHA256:             articleHash,
			Size:               int64(len(articleHTML)),
			Title:              document.Title,
			Byline:             document.Byline,
			SiteName:           document.SiteName,
			Language:           document.Language,
			Excerpt:            document.Excerpt,
			PublishedAt:        document.PublishedAt,
			ModifiedAt:         document.ModifiedAt,
			WordCount:          document.WordCount,
			ReadingTimeMinutes: document.ReadingTimeMinutes,
			LeadImageURL:       document.LeadImageURL,
			AuthorCandidates:   document.AuthorCandidates,
			Blocks:             document.Blocks,
			Outline:            document.Outline,
		},
		Resources: resources,
		Warnings:  document.Warnings,
	}
}

func writeFile(root, relativePath string, data []byte) error {
	path := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create directory for %s: %w", relativePath, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", relativePath, err)
	}
	return nil
}

func checksum(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func imageExtension(contentType string) string {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/avif":
		return ".avif"
	case "image/bmp":
		return ".bmp"
	case "image/gif":
		return ".gif"
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/svg+xml":
		return ".svg"
	case "image/tiff":
		return ".tiff"
	case "image/vnd.microsoft.icon", "image/x-icon":
		return ".ico"
	case "image/webp":
		return ".webp"
	default:
		return ".img"
	}
}
