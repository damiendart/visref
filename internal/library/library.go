package library

import (
	"context"
	"github.com/google/uuid"
	"io"
	"time"
)

// Item is an item in a visual reference library.
type Item struct {
	ID               uuid.UUID
	AlternativeText  string
	Description      string
	MimeType         string
	OriginalFilename string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ItemRepository is the interface that must be implemented for
// storing, retrieving, and deleting visual reference library items.
type ItemRepository interface {
	Create(ctx context.Context, item *Item, file io.Reader) error
	Get(ctx context.Context, id uuid.UUID) (*Item, error)
}
