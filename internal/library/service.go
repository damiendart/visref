// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package library

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"image"
	"image/color"
	// The following package is only imported for their side effect of
	// adding support for decoding GIF images.
	_ "image/gif"
	"image/jpeg"
	// The following package is only imported for their side effect of
	// adding support for decoding PNG images.
	_ "image/png"
	"io"
	"math"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"uuid"

	"golang.org/x/image/draw"
	// The following package is only imported for the side effect of
	// adding support for decoding WebP images.
	_ "golang.org/x/image/webp"

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
	Width            int
	Height           int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Service represents a service for managing visual reference library
// items. Files are stored on the local filesystem and metadata is
// stored in an accompanying SQLite database.
type Service struct {
	db             *sqlite.DB
	mediaRoot      *os.Root
	thumbnailsRoot *os.Root
}

// ErrNotFound is returned if a library item cannot be found.
var ErrNotFound = errors.New("item not found")

// ErrInvalidThumbnailFilename is returned if the thumbnail filename is invalid.
var ErrInvalidThumbnailFilename = errors.New("invalid thumbnail filename")

// NewService returns a new [Service].
func NewService(db *sqlite.DB, mediaRoot *os.Root, thumbnailsRoot *os.Root) *Service {
	return &Service{db, mediaRoot, thumbnailsRoot}
}

// CreateItem stores a new [Item].
func (s *Service) CreateItem(ctx context.Context, item *Item, file io.Reader) error {
	now := s.db.Now()
	u := uuid.NewV7()

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

	if _, err = dst.Seek(0, 0); err != nil {
		return err
	}

	config, _, err := image.DecodeConfig(dst)
	if err != nil {
		return err
	}

	if _, err = s.db.ExecContext(
		ctx,
		`INSERT INTO items (id, alternative_text, source, description, media_type, filepath, original_filename, width, height, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
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
		config.Width,
		config.Height,
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

// DeleteItemByID deletes a library item.
func (s *Service) DeleteItemByID(ctx context.Context, id uuid.UUID) error {
	row := s.db.QueryRowContext(
		ctx,
		`DELETE FROM items WHERE id = ? RETURNING filepath`,
		id.String(),
	)

	var p string

	if err := row.Scan(&p); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("item %s: %w", id.String(), ErrNotFound)
		}
	}

	if err := s.mediaRoot.Remove(p); err != nil {
		return err
	}

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
			width,
			height,
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
		&item.Width,
		&item.Height,
		(*sqlite.NullTime)(&item.CreatedAt),
		(*sqlite.NullTime)(&item.UpdatedAt),
	); err != nil {
		return nil, err
	}

	return &item, nil
}

// GetOriginalFileByItem returns the original uploaded file for an item.
func (s *Service) GetOriginalFileByItem(item *Item) (io.ReadSeekCloser, error) {
	f, err := s.mediaRoot.Open(item.Filepath)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// GetThumbnail returns a thumbnail for an item. The item and thumbnail
// width are inferred from the thumbnail filename.
func (s *Service) GetThumbnail(ctx context.Context, filename string) (io.ReadSeekCloser, time.Time, error) {
	if len(filename) < 4 {
		return nil, time.Time{}, ErrInvalidThumbnailFilename
	}

	p := filepath.Join(filename[0:2], filename[2:4], filename+".jpg")

	thumbnail, err := s.thumbnailsRoot.Open(p)
	if err == nil {
		stat, err := thumbnail.Stat()
		if err != nil {
			return nil, time.Time{}, err
		}

		return thumbnail, stat.ModTime(), err
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, time.Time{}, err
	}

	parts := strings.Split(filename, "--")
	if len(parts) != 2 {
		return nil, time.Time{}, ErrInvalidThumbnailFilename
	}

	id, err := uuid.Parse(parts[0])
	if err != nil {
		return nil, time.Time{}, ErrInvalidThumbnailFilename
	}

	width, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, time.Time{}, ErrInvalidThumbnailFilename
	}

	err = s.createThumbnail(ctx, id, width)
	if err != nil {
		return nil, time.Time{}, err
	}

	thumbnail, err = s.thumbnailsRoot.Open(p)
	if err != nil {
		return nil, time.Time{}, err
	}

	stat, err := thumbnail.Stat()
	if err != nil {
		return nil, time.Time{}, err
	}

	return thumbnail, stat.ModTime(), err
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

func (s *Service) createThumbnail(ctx context.Context, id uuid.UUID, width int) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	item, err := s.GetItemByID(context.TODO(), id)
	if err != nil {
		return ErrNotFound
	}

	f, err := s.GetOriginalFileByItem(item)
	if err != nil {
		return err
	}
	defer f.Close()

	i, _, err := image.Decode(f)
	if err != nil {
		return err
	}

	dir := filepath.Join(item.ID.String()[0:2], item.ID.String()[2:4])

	if err := s.thumbnailsRoot.MkdirAll(dir, 0700); err != nil {
		return err
	}

	thumbnail, err := s.thumbnailsRoot.Create(
		filepath.Join(dir, fmt.Sprintf("%s--%d.jpg", item.ID, width)),
	)
	if err != nil {
		return err
	}
	defer thumbnail.Close()

	ratio := (float64)(i.Bounds().Max.Y) / (float64)(i.Bounds().Max.X)
	height := int(math.Round(float64(width) * ratio))

	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	draw.Draw(dst, dst.Bounds(), &image.Uniform{C: color.RGBA{R: 255, G: 255, B: 255, A: 1.0}}, image.Point{}, draw.Src)
	draw.ApproxBiLinear.Scale(dst, dst.Rect, i, i.Bounds(), draw.Over, nil)

	return jpeg.Encode(thumbnail, dst, &jpeg.Options{Quality: jpeg.DefaultQuality})
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
	case "image/webp":
		return ".webp", nil
	}

	return "", errors.New("media type not supported")
}
