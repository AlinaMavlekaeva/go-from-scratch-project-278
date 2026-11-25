package links

import (
	"code/internal/db"
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"os"

	"github.com/jackc/pgx/v5/pgconn"
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

func generateShortURL(shortName string) (string, error) {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		return "", ErrEmptyBaseURL
	}
	shortURL := baseURL + "/r/" + shortName
	return shortURL, nil
}

func GetLinks(ctx context.Context, conn db.DBTX) ([]db.GetLinksRow, error) {
	q := db.New(conn)
	links, err := q.GetLinks(ctx)
	return links, err
}

func GetLinksPage(ctx context.Context, conn db.DBTX, limit int, offset int) ([]db.GetLinksPageRow, int64, error) {
	q := db.New(conn)
	links, err := q.GetLinksPage(ctx, db.GetLinksPageParams{Offset: int32(offset), Limit: int32(limit)})
	if err != nil {
		return []db.GetLinksPageRow{}, 0, err
	}
	count, err := q.CountLinks(ctx)
	return links, count, err
}

func CreateLink(ctx context.Context, conn db.DBTX, params CreateLinkParams) (db.CreateLinkRow, error) {
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
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "links_short_name_key" {
			return db.CreateLinkRow{}, ErrShortNameConflict
		}
	}
	return link, err
}

func UpdateLinkByID(ctx context.Context, conn db.DBTX, params UpdateLinkParams) (db.UpdateLinkByIDRow, error) {
	if params.ID <= 0 {
		return db.UpdateLinkByIDRow{}, ErrInvalidID
	}
	q := db.New(conn)
	var shortUrl string
	if params.ShortName != "" {
		var err error
		shortUrl, err = generateShortURL(params.ShortName)
		if err != nil {
			return db.UpdateLinkByIDRow{}, fmt.Errorf("update link error: %w", err)
		}
	}
	link, err := q.UpdateLinkByID(ctx, db.UpdateLinkByIDParams{ID: params.ID,
		OriginalUrl: params.OriginalUrl,
		ShortName:   params.ShortName,
		ShortUrl:    shortUrl})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.UpdateLinkByIDRow{}, fmt.Errorf("update link error: %w", ErrNotFound)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "links_short_name_key" {
			return db.UpdateLinkByIDRow{}, ErrShortNameConflict
		}
		return db.UpdateLinkByIDRow{}, fmt.Errorf("update link error: %w", err)
	}
	return link, nil
}

func GetLinkByID(ctx context.Context, conn db.DBTX, id int64) (db.GetLinkByIDRow, error) {
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

func GetLinkByShortName(ctx context.Context, conn db.DBTX, shortName string) (db.GetLinkByShortNameRow, error) {
	if shortName == "" {
		return db.GetLinkByShortNameRow{}, fmt.Errorf("get link error: %w", ErrEmptyShortName)
	}
	q := db.New(conn)
	link, err := q.GetLinkByShortName(ctx, shortName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.GetLinkByShortNameRow{}, fmt.Errorf("get link error: %w", ErrNotFound)
		}
		return db.GetLinkByShortNameRow{}, fmt.Errorf("get link error: %w", err)
	}
	return link, nil
}

func DeleteLink(ctx context.Context, conn db.DBTX, id int64) error {
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

func GetLinkVisits(ctx context.Context, conn db.DBTX) ([]db.GetLinkVisitsRow, error) {
	q := db.New(conn)
	linkVisits, err := q.GetLinkVisits(ctx)
	return linkVisits, err
}

func GetLinkVisitsPage(ctx context.Context, conn db.DBTX, limit int, offset int) ([]db.GetLinkVisitsPageRow, int64, error) {
	q := db.New(conn)
	linkVisits, err := q.GetLinkVisitsPage(ctx, db.GetLinkVisitsPageParams{Offset: int32(offset), Limit: int32(limit)})
	if err != nil {
		return []db.GetLinkVisitsPageRow{}, 0, err
	}
	count, err := q.CountLinkVisits(ctx)
	return linkVisits, count, err
}

func CreateLinkVisit(ctx context.Context, conn db.DBTX, params db.CreateLinkVisitParams) (db.CreateLinkVisitRow, error) {
	q := db.New(conn)
	linkVisit, err := q.CreateLinkVisit(ctx, db.CreateLinkVisitParams{LinkID: params.LinkID, Ip: params.Ip, UserAgent: params.UserAgent, Referer: params.Referer, Status: params.Status})
	return linkVisit, err
}
