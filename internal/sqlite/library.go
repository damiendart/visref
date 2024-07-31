package sqlite

import (
	"context"
	"fmt"
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
func (r *ItemRepository) Create(ctx context.Context, item *library.Item) error {
	tx, err := r.db.BeginTx(ctx, nil)
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
		`INSERT INTO items (id, title, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		u,
		item.Title,
		item.Description,
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

	rows, err := tx.QueryContext(ctx, `SELECT id, title, description, created_at, updated_at FROM items WHERE id = ?`, id.String())
	if err != nil {
		return nil, err
	}

	items := make([]*library.Item, 0)

	for rows.Next() {
		var item library.Item

		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Description,
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
