package sqlite

import (
	"context"
	"github.com/damiendart/visref/internal/library"
	"github.com/google/uuid"
)

// ItemRepository is an implementation of library.ItemRepository which
// uses SQLite to store library items.
type ItemRepository struct {
	db *DB
}

// NewItemRepository returns a new ItemRepository.
func NewItemRepository(db *DB) *ItemRepository {
	return &ItemRepository{db}
}

// Create stores a new library.Item in the database.
func (s *ItemRepository) Create(ctx context.Context, item *library.Item) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	now := tx.now

	u, err := uuid.NewV7()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO media (id, title, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		u,
		item.Title,
		item.Description,
		now,
		now,
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
