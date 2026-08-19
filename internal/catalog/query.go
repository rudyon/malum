package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/rudyon/malum/internal/identifier"
)

const documentSelect = `
	SELECT
		d.id, d.reading_kind, d.acquisition_method, d.original_format,
		d.source_url, d.title, d.description, d.site_name, d.raw_byline,
		d.language, d.published_at, d.source_modified_at,
		d.word_count, d.reading_time_minutes, d.thumbnail_path, d.saved_at,
		a.id, a.handle, a.display_name, a.avatar_source_url, a.avatar_path
	FROM documents d
	LEFT JOIN authors a ON a.id = d.author_id`

func (c *Catalog) ListDocuments(ctx context.Context) ([]Document, error) {
	rows, err := c.database.QueryContext(ctx, documentSelect+" ORDER BY d.saved_at DESC, d.id")
	if err != nil {
		return nil, fmt.Errorf("list catalogue documents: %w", err)
	}
	defer rows.Close()

	documents := make([]Document, 0)
	for rows.Next() {
		document, err := scanDocument(rows)
		if err != nil {
			return nil, err
		}
		documents = append(documents, document)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read catalogue documents: %w", err)
	}
	return documents, nil
}

func (c *Catalog) GetDocument(ctx context.Context, documentID string) (Document, error) {
	document, err := scanDocument(c.database.QueryRowContext(ctx, documentSelect+" WHERE d.id = ?", documentID))
	if errors.Is(err, sql.ErrNoRows) {
		return Document{}, fmt.Errorf("%w: %s", ErrDocumentNotFound, documentID)
	}
	return document, err
}

func (c *Catalog) GetAuthor(ctx context.Context, authorID string) (Author, error) {
	if !identifier.IsUUID(authorID) {
		return Author{}, fmt.Errorf("get author: invalid UUID %q", authorID)
	}
	var author Author
	var avatarSource, avatarPath sql.NullString
	err := c.database.QueryRowContext(ctx, `
		SELECT id, handle, display_name, avatar_source_url, avatar_path
		FROM authors WHERE id = ?`, authorID,
	).Scan(&author.ID, &author.Handle, &author.DisplayName, &avatarSource, &avatarPath)
	if errors.Is(err, sql.ErrNoRows) {
		return Author{}, fmt.Errorf("%w: %s", ErrAuthorNotFound, authorID)
	}
	if err != nil {
		return Author{}, fmt.Errorf("get author: %w", err)
	}
	author.AvatarSourceURL = avatarSource.String
	author.AvatarPath = avatarPath.String
	return author, nil
}

func (c *Catalog) SetAuthorAvatar(ctx context.Context, authorID, sourceURL, relativePath string) error {
	if !identifier.IsUUID(authorID) {
		return fmt.Errorf("set author avatar: invalid author UUID %q", authorID)
	}
	cleanPath := filepath.ToSlash(filepath.Clean(relativePath))
	expectedPrefix := filepath.ToSlash(filepath.Join("authors", authorID, "avatars")) + "/"
	if filepath.IsAbs(relativePath) || !strings.HasPrefix(cleanPath, expectedPrefix) {
		return errors.New("set author avatar: path must be inside the author's avatar directory")
	}
	result, err := c.database.ExecContext(ctx, `
		UPDATE authors SET avatar_source_url = ?, avatar_path = ? WHERE id = ?`,
		nullableString(sourceURL), cleanPath, authorID,
	)
	if err != nil {
		return fmt.Errorf("set author avatar: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count updated authors: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("%w: %s", ErrAuthorNotFound, authorID)
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanDocument(scanner rowScanner) (Document, error) {
	var document Document
	var language, published, modified, thumbnail sql.NullString
	var authorID, authorHandle, authorName, avatarSource, avatarPath sql.NullString
	var saved string
	err := scanner.Scan(
		&document.ID,
		&document.ReadingKind,
		&document.AcquisitionMethod,
		&document.OriginalFormat,
		&document.SourceURL,
		&document.Title,
		&document.Description,
		&document.SiteName,
		&document.RawByline,
		&language,
		&published,
		&modified,
		&document.WordCount,
		&document.ReadingTimeMinutes,
		&thumbnail,
		&saved,
		&authorID,
		&authorHandle,
		&authorName,
		&avatarSource,
		&avatarPath,
	)
	if err != nil {
		return Document{}, err
	}
	document.Language = language.String
	document.ThumbnailPath = thumbnail.String
	savedAt, parseErr := parseTime(saved)
	if parseErr != nil {
		return Document{}, fmt.Errorf("parse saved time for document %s: %w", document.ID, parseErr)
	}
	document.SavedAt = savedAt
	if published.Valid {
		value, err := parseTime(published.String)
		if err != nil {
			return Document{}, fmt.Errorf("parse publication time for document %s: %w", document.ID, err)
		}
		document.PublishedAt = &value
	}
	if modified.Valid {
		value, err := parseTime(modified.String)
		if err != nil {
			return Document{}, fmt.Errorf("parse modified time for document %s: %w", document.ID, err)
		}
		document.SourceModifiedAt = &value
	}
	if authorID.Valid {
		document.Author = &Author{
			ID:              authorID.String,
			Handle:          authorHandle.String,
			DisplayName:     authorName.String,
			AvatarSourceURL: avatarSource.String,
			AvatarPath:      avatarPath.String,
		}
	}
	return document, nil
}

func parseTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, err
	}
	return parsed.UTC(), nil
}
