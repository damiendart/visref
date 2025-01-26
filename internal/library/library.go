package library

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

// Item is an item in a visual reference library.
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

// ItemRepository is the interface that must be implemented for
// storing, retrieving, and deleting visual reference library items.
type ItemRepository interface {
	Create(ctx context.Context, item *Item, file io.Reader) error
	Get(ctx context.Context, id uuid.UUID) (*Item, error)
}
