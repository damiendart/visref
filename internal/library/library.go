package library

import (
	"context"
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
	Description      string
	MimeType         string
	Filepath         string
	OriginalFilename string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Service represents a service for managing visual reference library
// items. Files are stored on the local filesystem and metadata is
// stored in an accompanying SQLite database.
type Service struct {
	db       *sqlite.DB
	mediaDir string
}

// NewService returns a new [Service].
func NewService(db *sqlite.DB, mediaDir string) *Service {
	return &Service{db, mediaDir}
}

// CreateItem stores a new [Item].
func (s *Service) CreateItem(ctx context.Context, item *Item, file io.Reader) error {
	u, err := uuid.NewV7()
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	now := tx.Now

	ext, err := mime.ExtensionsByType(item.MimeType)
	if err != nil {
		return err
	}

	err = os.MkdirAll(
		filepath.Join(s.mediaDir, now.Format("2006/01")),
		0700,
	)
	if err != nil {
		return err
	}

	dst, err := os.Create(
		filepath.Join(
			s.mediaDir,
			now.Format("2006/01"),
			fmt.Sprintf("%s%s", u.String(), ext[0]),
		),
	)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO items (id, alternative_text, description, mime_type, filepath, original_filename, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		u,
		item.AlternativeText,
		item.Description,
		item.MimeType,
		filepath.Join(
			now.Format("2006/01"),
			fmt.Sprintf("%s%s", u.String(), ext[0]),
		),
		item.OriginalFilename,
		(*sqlite.NullTime)(&now),
		(*sqlite.NullTime)(&now),
	)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return rollbackErr
		}

		return err
	}

	item.ID = u
	item.CreatedAt = now
	item.UpdatedAt = now

	return nil
}

// GetItemByID retrieves a library item by ID.
func (s *Service) GetItemByID(ctx context.Context, id uuid.UUID) (*Item, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, `SELECT id, alternative_text, description, mime_type, filepath, original_filename, created_at, updated_at FROM items WHERE id = ?`, id.String())
	if err != nil {
		return nil, err
	}

	items := make([]*Item, 0)

	for rows.Next() {
		var item Item

		if err := rows.Scan(
			&item.ID,
			&item.AlternativeText,
			&item.Description,
			&item.MimeType,
			&item.Filepath,
			&item.OriginalFilename,
			(*sqlite.NullTime)(&item.CreatedAt),
			(*sqlite.NullTime)(&item.UpdatedAt),
		); err != nil {
			return nil, err
		}

		items = append(items, &item)
	}

	err = rows.Close()
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("unable to find item %q", id)
	}

	return items[0], nil
}

// GetOriginalFileByItem returns the original uploaded file for an item.
func (s *Service) GetOriginalFileByItem(item *Item) (io.ReadSeeker, error) {
	f, err := os.Open(filepath.Join(s.mediaDir, item.Filepath))
	if err != nil {
		return nil, err
	}

	return f, nil
}
