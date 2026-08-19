package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rudyon/malum/internal/identifier"
	"github.com/rudyon/malum/internal/ingest/webpage"
	documentstore "github.com/rudyon/malum/internal/storage/document"
)

func (c *Catalog) AddDocument(ctx context.Context, saved documentstore.Saved) (Document, error) {
	manifest := saved.Manifest
	if err := validateSavedDocument(saved); err != nil {
		return Document{}, err
	}

	transaction, err := c.database.BeginTx(ctx, nil)
	if err != nil {
		return Document{}, fmt.Errorf("begin catalogue publication: %w", err)
	}
	defer func() { _ = transaction.Rollback() }()

	var exists int
	err = transaction.QueryRowContext(ctx, "SELECT 1 FROM documents WHERE id = ?", saved.ID).Scan(&exists)
	if err == nil {
		return Document{}, fmt.Errorf("%w: %s", ErrDocumentExists, saved.ID)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Document{}, fmt.Errorf("check catalogue document: %w", err)
	}

	author, err := c.resolvePrimaryAuthor(ctx, transaction, manifest.Article.AuthorCandidates, manifest.StoredAt)
	if err != nil {
		return Document{}, err
	}
	var authorID any
	if author != nil {
		authorID = author.ID
	}

	_, err = transaction.ExecContext(ctx, `
		INSERT INTO documents (
			id, reading_kind, acquisition_method, original_format, source_url,
			title, description, site_name, raw_byline, author_id, language,
			published_at, source_modified_at, word_count, reading_time_minutes,
			thumbnail_path, saved_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		saved.ID,
		manifest.ReadingKind,
		manifest.Acquisition.Method,
		manifest.Original.Format,
		manifest.Acquisition.FinalURL,
		manifest.Article.Title,
		manifest.Article.Excerpt,
		manifest.Article.SiteName,
		manifest.Article.Byline,
		authorID,
		nullableString(manifest.Article.Language),
		nullableTime(manifest.Article.PublishedAt),
		nullableTime(manifest.Article.ModifiedAt),
		manifest.Article.WordCount,
		manifest.Article.ReadingTimeMinutes,
		nullableString(thumbnailPath(manifest)),
		formatTime(manifest.StoredAt),
	)
	if err != nil {
		return Document{}, fmt.Errorf("insert catalogue document: %w", err)
	}
	if err := transaction.Commit(); err != nil {
		return Document{}, fmt.Errorf("commit catalogue publication: %w", err)
	}

	return documentFromManifest(manifest, author), nil
}

func validateSavedDocument(saved documentstore.Saved) error {
	manifest := saved.Manifest
	if !identifier.IsUUID(saved.ID) || saved.ID != manifest.DocumentID {
		return errors.New("publish document: saved ID and manifest ID must be the same valid UUID")
	}
	if manifest.SchemaVersion != documentstore.SchemaVersion {
		return fmt.Errorf("publish document: unsupported bundle schema version %d", manifest.SchemaVersion)
	}
	if strings.TrimSpace(manifest.Article.Title) == "" || strings.TrimSpace(manifest.Acquisition.FinalURL) == "" {
		return errors.New("publish document: title and final source URL are required")
	}
	return nil
}

func (c *Catalog) resolvePrimaryAuthor(ctx context.Context, transaction *sql.Tx, candidates []webpage.AuthorCandidate, createdAt time.Time) (*Author, error) {
	if len(candidates) == 0 || strings.TrimSpace(candidates[0].Name) == "" {
		return nil, nil
	}
	candidate := candidates[0]
	candidate.Name = strings.TrimSpace(candidate.Name)

	authorID, err := matchingIdentityAuthor(ctx, transaction, candidate.Identities)
	if err != nil {
		return nil, err
	}
	if authorID == "" {
		authorID, err = uniqueNameAuthor(ctx, transaction, comparisonKey(candidate.Name))
		if err != nil {
			return nil, err
		}
	}

	if authorID == "" {
		authorID, err = c.createAuthor(ctx, transaction, candidate, createdAt)
		if err != nil {
			return nil, err
		}
	}
	if err := addAuthorEvidence(ctx, transaction, authorID, candidate); err != nil {
		return nil, err
	}
	return loadAuthor(ctx, transaction, authorID)
}

func matchingIdentityAuthor(ctx context.Context, transaction *sql.Tx, identities []webpage.AuthorIdentity) (string, error) {
	matched := ""
	for _, identity := range identities {
		kind := strings.TrimSpace(identity.Kind)
		valueKey := identityKey(kind, identity.Value)
		if kind == "" || valueKey == "" {
			continue
		}
		var authorID string
		err := transaction.QueryRowContext(ctx,
			"SELECT author_id FROM author_identities WHERE kind = ? AND value_key = ?",
			kind, valueKey,
		).Scan(&authorID)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("match author identity: %w", err)
		}
		if matched != "" && matched != authorID {
			return "", errors.New("resolve author: structured identities refer to different existing authors")
		}
		matched = authorID
	}
	return matched, nil
}

func uniqueNameAuthor(ctx context.Context, transaction *sql.Tx, nameKey string) (string, error) {
	if nameKey == "" {
		return "", nil
	}
	rows, err := transaction.QueryContext(ctx, "SELECT id FROM authors WHERE name_key = ? ORDER BY id LIMIT 2", nameKey)
	if err != nil {
		return "", fmt.Errorf("match author name: %w", err)
	}
	defer rows.Close()
	var matches []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return "", fmt.Errorf("scan matching author: %w", err)
		}
		matches = append(matches, id)
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("read matching authors: %w", err)
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	return "", nil
}

func (c *Catalog) createAuthor(ctx context.Context, transaction *sql.Tx, candidate webpage.AuthorCandidate, createdAt time.Time) (string, error) {
	authorID, err := c.newID()
	if err != nil {
		return "", fmt.Errorf("generate author ID: %w", err)
	}
	if !identifier.IsUUID(authorID) {
		return "", fmt.Errorf("generate author ID: invalid UUID %q", authorID)
	}
	handle, err := availableHandle(ctx, transaction, generatedHandle(candidate.Name))
	if err != nil {
		return "", err
	}
	_, err = transaction.ExecContext(ctx, `
		INSERT INTO authors (
			id, handle, handle_key, display_name, name_key,
			avatar_source_url, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		authorID,
		handle,
		comparisonKey(handle),
		candidate.Name,
		comparisonKey(candidate.Name),
		nullableString(candidate.ImageURL),
		formatTime(createdAt),
	)
	if err != nil {
		return "", fmt.Errorf("insert catalogue author: %w", err)
	}
	return authorID, nil
}

func availableHandle(ctx context.Context, transaction *sql.Tx, base string) (string, error) {
	for suffix := 1; ; suffix++ {
		candidate := base
		if suffix > 1 {
			candidate = fmt.Sprintf("%s%d", base, suffix)
		}
		var exists int
		err := transaction.QueryRowContext(ctx, "SELECT 1 FROM authors WHERE handle_key = ?", comparisonKey(candidate)).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return candidate, nil
		}
		if err != nil {
			return "", fmt.Errorf("check author handle: %w", err)
		}
	}
}

func addAuthorEvidence(ctx context.Context, transaction *sql.Tx, authorID string, candidate webpage.AuthorCandidate) error {
	for _, identity := range candidate.Identities {
		kind := strings.TrimSpace(identity.Kind)
		value := strings.TrimSpace(identity.Value)
		valueKey := identityKey(kind, value)
		if kind == "" || value == "" || valueKey == "" {
			continue
		}
		result, err := transaction.ExecContext(ctx, `
			INSERT INTO author_identities (author_id, kind, value, value_key)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(kind, value_key) DO NOTHING`,
			authorID, kind, value, valueKey,
		)
		if err != nil {
			return fmt.Errorf("insert author identity: %w", err)
		}
		if affected, _ := result.RowsAffected(); affected == 0 {
			var existing string
			if err := transaction.QueryRowContext(ctx,
				"SELECT author_id FROM author_identities WHERE kind = ? AND value_key = ?",
				kind, valueKey,
			).Scan(&existing); err != nil {
				return fmt.Errorf("verify author identity: %w", err)
			}
			if existing != authorID {
				return errors.New("resolve author: identity belongs to a different author")
			}
		}
	}
	if candidate.ImageURL != "" {
		if _, err := transaction.ExecContext(ctx, `
			UPDATE authors
			SET avatar_source_url = CASE
				WHEN avatar_source_url IS NULL OR avatar_source_url = '' THEN ?
				ELSE avatar_source_url
			END
			WHERE id = ?`, candidate.ImageURL, authorID); err != nil {
			return fmt.Errorf("update author avatar source: %w", err)
		}
	}
	return nil
}

func loadAuthor(ctx context.Context, transaction *sql.Tx, authorID string) (*Author, error) {
	var author Author
	var avatarSource, avatarPath sql.NullString
	if err := transaction.QueryRowContext(ctx, `
		SELECT id, handle, display_name, avatar_source_url, avatar_path
		FROM authors WHERE id = ?`, authorID,
	).Scan(&author.ID, &author.Handle, &author.DisplayName, &avatarSource, &avatarPath); err != nil {
		return nil, fmt.Errorf("load catalogue author: %w", err)
	}
	author.AvatarSourceURL = avatarSource.String
	author.AvatarPath = avatarPath.String
	return &author, nil
}

func thumbnailPath(manifest documentstore.Manifest) string {
	for _, resource := range manifest.Resources {
		if resource.Status == documentstore.ResourceStored && resource.Role == "lead-image" {
			return resource.Path
		}
	}
	return ""
}

func documentFromManifest(manifest documentstore.Manifest, author *Author) Document {
	return Document{
		ID:                 manifest.DocumentID,
		ReadingKind:        manifest.ReadingKind,
		AcquisitionMethod:  manifest.Acquisition.Method,
		OriginalFormat:     manifest.Original.Format,
		SourceURL:          manifest.Acquisition.FinalURL,
		Title:              manifest.Article.Title,
		Description:        manifest.Article.Excerpt,
		SiteName:           manifest.Article.SiteName,
		RawByline:          manifest.Article.Byline,
		Author:             author,
		Language:           manifest.Article.Language,
		PublishedAt:        manifest.Article.PublishedAt,
		SourceModifiedAt:   manifest.Article.ModifiedAt,
		WordCount:          manifest.Article.WordCount,
		ReadingTimeMinutes: manifest.Article.ReadingTimeMinutes,
		ThumbnailPath:      thumbnailPath(manifest),
		SavedAt:            manifest.StoredAt.UTC(),
	}
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return formatTime(*value)
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}
