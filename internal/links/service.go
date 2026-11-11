package links

import (
	"code/internal/db"
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"os"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateName10() (string, error) {
	result := make([]byte, 10)
	for i := range result {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", fmt.Errorf("generate short name error: %w", err)
		}
		result[i] = alphabet[index.Int64()]
	}
	return string(result), nil
}

type CreateLinkParams struct {
	OriginalUrl string
	ShortName   string
}

type UpdateLinkParams struct {
	ID          int64
	OriginalUrl string
	ShortName   string
}

func validateURL(rawURL string) error {
	if rawURL != "" {
		parsed, err := url.Parse(rawURL)
		if err != nil {
			return fmt.Errorf("invalid URL: %w", err)
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return fmt.Errorf("invalid URL: %w", ErrUnsupportScheme)
		}
		if parsed.Hostname() != "" {
			return nil
		}
		return fmt.Errorf("invalid URL: %w", ErrEmptyHostname)
	}
	return fmt.Errorf("invalid URL: %w", ErrEmptyURL)
}

func generateShortURL(shortName string) (string, error) {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		return "", ErrEmptyBaseURL
	}
	shortURL := baseURL + "/" + shortName
	return shortURL, nil
}

func GetLinks(ctx context.Context, conn *sql.DB) ([]db.GetLinksRow, error) {
	q := db.New(conn)
	links, err := q.GetLinks(ctx)
	return links, err
}

func CreateLink(ctx context.Context, conn *sql.DB, params CreateLinkParams) (db.CreateLinkRow, error) {
	if err := validateURL(params.OriginalUrl); err != nil {
		return db.CreateLinkRow{}, fmt.Errorf("create link error: %w", errors.Join(ErrValidate, err))
	}
	if params.ShortName == "" {
		name, err := generateName10()
		if err != nil {
			return db.CreateLinkRow{}, fmt.Errorf("create link error: %w", err)
		}
		params.ShortName = name
	}
	shortURL, err := generateShortURL(params.ShortName)
	if err != nil {
		return db.CreateLinkRow{}, fmt.Errorf("create link error: %w", err)
	}
	q := db.New(conn)
	link, err := q.CreateLink(ctx, db.CreateLinkParams{OriginalUrl: params.OriginalUrl, ShortName: params.ShortName, ShortUrl: shortURL})
	return link, err
}

func UpdateLinkByID(ctx context.Context, conn *sql.DB, params UpdateLinkParams) (db.UpdateLinkByIDRow, error) {
	if params.ID <= 0 {
		return db.UpdateLinkByIDRow{}, fmt.Errorf("update link error: %w", errors.Join(ErrValidate, ErrInvalidID))
	}
	if err := validateURL(params.OriginalUrl); err != nil {
		return db.UpdateLinkByIDRow{}, fmt.Errorf("update link error: %w", errors.Join(ErrValidate, err))
	}
	if params.ShortName == "" {
		return db.UpdateLinkByIDRow{}, fmt.Errorf("update link error: %w", ErrEmptyShortName)
	}
	shortURL, err := generateShortURL(params.ShortName)
	if err != nil {
		return db.UpdateLinkByIDRow{}, fmt.Errorf("update link error: %w", err)
	}
	q := db.New(conn)
	link, err := q.UpdateLinkByID(ctx, db.UpdateLinkByIDParams{ID: params.ID,
		OriginalUrl: params.OriginalUrl,
		ShortName:   params.ShortName,
		ShortUrl:    shortURL})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.UpdateLinkByIDRow{}, fmt.Errorf("update link error: %w", ErrNotFound)
		}
		return db.UpdateLinkByIDRow{}, fmt.Errorf("update link error: %w", err)
	}
	return link, nil
}

func GetLinkByID(ctx context.Context, conn *sql.DB, id int64) (db.GetLinkByIDRow, error) {
	if id <= 0 {
		return db.GetLinkByIDRow{}, fmt.Errorf("get link error: %w", ErrInvalidID)
	}
	q := db.New(conn)
	link, err := q.GetLinkByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.GetLinkByIDRow{}, fmt.Errorf("get link error: %w", ErrNotFound)
		}
		return db.GetLinkByIDRow{}, fmt.Errorf("get link error: %w", err)
	}
	return link, nil
}

func DeleteLink(ctx context.Context, conn *sql.DB, id int64) error {
	if id <= 0 {
		return fmt.Errorf("delete link error: %w", ErrInvalidID)
	}
	q := db.New(conn)
	err := q.DeleteLink(ctx, id)
	if err != nil {
		return fmt.Errorf("delete link error: %w", err)
	}
	return nil
}
