// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package library

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/damiendart/visref/internal/sqlite"
)

// Item represents an item from a visual reference library.
type Item struct {
	ID               uuid.UUID
	AlternativeText  string
	Source           string
	Description      string
	MediaType        string
	Filepath         string
	OriginalFilename string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Service represents a service for managing visual reference library
// items. Files are stored on the local filesystem and metadata is
// stored in an accompanying SQLite database.
type Service struct {
	db        *sqlite.DB
	mediaRoot *os.Root
}

// NewService returns a new [Service].
func NewService(db *sqlite.DB, mediaRoot *os.Root) *Service {
	return &Service{db, mediaRoot}
}

// CreateItem stores a new [Item].
func (s *Service) CreateItem(ctx context.Context, item *Item, file io.Reader) error {
	now := s.db.Now()

	u, err := uuid.NewV7()
	if err != nil {
		return err
	}

	ext, err := getExtensionByMediaType(item.MediaType)
	if err != nil {
		return err
	}

	if err := s.mediaRoot.MkdirAll(now.Format("2006/01"), 0700); err != nil {
		return err
	}

	dst, err := s.mediaRoot.Create(
		filepath.Join(
			now.Format("2006/01"),
			fmt.Sprintf("%s%s", u.String(), ext),
		),
	)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, file); err != nil {
		return err
	}

	if _, err = s.db.ExecContext(
		ctx,
		`INSERT INTO items (id, alternative_text, source, description, media_type, filepath, original_filename, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		u,
		item.AlternativeText,
		item.Source,
		item.Description,
		item.MediaType,
		filepath.Join(
			now.Format("2006/01"),
			fmt.Sprintf("%s%s", u.String(), ext),
		),
		item.OriginalFilename,
		(*sqlite.NullTime)(&now),
		(*sqlite.NullTime)(&now),
	); err != nil {
		return err
	}

	item.ID = u
	item.CreatedAt = now
	item.UpdatedAt = now

	return nil
}

// GetItemByID retrieves a library item by ID.
func (s *Service) GetItemByID(ctx context.Context, id uuid.UUID) (*Item, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT
			id,
			alternative_text,
			source,
			description,
			media_type,
			filepath,
			original_filename,
			created_at,
			updated_at
		FROM items
		WHERE id = ?`,
		id.String(),
	)

	var item Item

	if err := row.Scan(
		&item.ID,
		&item.AlternativeText,
		&item.Source,
		&item.Description,
		&item.MediaType,
		&item.Filepath,
		&item.OriginalFilename,
		(*sqlite.NullTime)(&item.CreatedAt),
		(*sqlite.NullTime)(&item.UpdatedAt),
	); err != nil {
		return nil, err
	}

	return &item, nil
}

// GetOriginalFileByItem returns the original uploaded file for an item.
func (s *Service) GetOriginalFileByItem(item *Item) (io.ReadSeeker, error) {
	f, err := s.mediaRoot.Open(item.Filepath)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// PatchItem updates a library item.
func (s *Service) PatchItem(
	ctx context.Context,
	item *Item,
	alternativeText string,
	source string,
	description string,
) error {
	row := s.db.QueryRowContext(
		ctx,
		`
			UPDATE items
			SET
				alternative_text = ?,
				source = ?,
				description = ?,
				updated_at = ?
			WHERE id = ?
			RETURNING alternative_text, source, description, updated_at`,
		alternativeText,
		source,
		description,
		(*sqlite.NullTime)(new(time.Now())),
		item.ID.String(),
	)

	if err := row.Scan(
		&item.AlternativeText,
		&item.Source,
		&item.Description,
		(*sqlite.NullTime)(&item.UpdatedAt),
	); err != nil {
		return err
	}

	return nil
}

// IsAcceptedMediaType reports whether the given media type is accepted
// by the visual reference library.
func IsAcceptedMediaType(mediaType string) bool {
	_, err := getExtensionByMediaType(mediaType)

	return err == nil
}

func getExtensionByMediaType(mediaType string) (string, error) {
	m, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return "", err
	}

	switch m {
	case "image/jpeg":
		return ".jpg", nil
	case "image/png":
		return ".png", nil
	}

	return "", errors.New("media type not supported")
}
