package library

import (
	"context"
	"github.com/google/uuid"
	"time"
)

// Item is an item in a visual reference library.
type Item struct {
	ID          uuid.UUID
	Title       string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ItemRepository is the interface that must be implemented for
// storing, retrieving, and deleting visual reference library items.
type ItemRepository interface {
	Create(ctx context.Context, item *Item) error
}
