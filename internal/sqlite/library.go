package sqlite

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/damiendart/visref/internal/library"
)

// ItemRepository is an implementation of library.ItemRepository which
// uses SQLite to store library items.
type ItemRepository struct {
	db       *DB
	mediaDir string
}

// NewItemRepository returns a new ItemRepository.
func NewItemRepository(db *DB, mediaDir string) *ItemRepository {
	return &ItemRepository{db, mediaDir}
}

// Create stores a new library.Item in the database.
func (r *ItemRepository) Create(ctx context.Context, item *library.Item, file io.Reader) error {
	u, err := uuid.NewV7()
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	now := tx.now

	ext, err := mime.ExtensionsByType(item.MimeType)
	if err != nil {
		return err
	}

	err = os.MkdirAll(
		filepath.Join(r.mediaDir, now.Format("2006/01")),
		0700,
	)
	if err != nil {
		return err
	}

	dst, err := os.Create(
		filepath.Join(
			r.mediaDir,
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
		(*NullTime)(&now),
		(*NullTime)(&now),
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

// Get retrieves a library.Item from the database.
func (r *ItemRepository) Get(ctx context.Context, id uuid.UUID) (*library.Item, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, `SELECT id, alternative_text, description, mime_type, filepath, original_filename, created_at, updated_at FROM items WHERE id = ?`, id.String())
	if err != nil {
		return nil, err
	}

	items := make([]*library.Item, 0)

	for rows.Next() {
		var item library.Item

		if err := rows.Scan(
			&item.ID,
			&item.AlternativeText,
			&item.Description,
			&item.MimeType,
			&item.Filepath,
			&item.OriginalFilename,
			(*NullTime)(&item.CreatedAt),
			(*NullTime)(&item.UpdatedAt),
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
