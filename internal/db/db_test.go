//go:build integration

package db

import (
	"code/db/migrations"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var conn *sql.DB

type link struct {
	OriginalUrl string
	ShortName   string
	ShortUrl    string
	CreatedAt   time.Time
}

type linkVisit struct {
	LinkID    int64
	Ip        string
	UserAgent string
	Referer   sql.NullString
	Status    int32
	CreatedAt time.Time
}

func TestMain(m *testing.M) {
	os.Exit(runIntegrationTests(m))
}

func runIntegrationTests(m *testing.M) int {
	ctx := context.Background()

	pg, err := postgres.Run(ctx,
		"postgres:18",
		postgres.WithDatabase("links_app_test"),
		postgres.WithUsername("test_user"),
		postgres.WithPassword("test_password"),
		postgres.BasicWaitStrategies(),
	)
	if pg != nil {
		defer func() {
			err := pg.Terminate(ctx)
			if err != nil {
				fmt.Println(err)
			}
		}()
	}
	if err != nil {
		fmt.Println(err)
		return 1
	}
	databaseURL, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Println("database connection error: ", err)
		return 1
	}
	conn, err = sql.Open("pgx", databaseURL)
	if err != nil {
		fmt.Println("database initialization error: ", err)
		return 1
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Println("database close error: ", err)
		}
	}()
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	err = conn.PingContext(pingCtx)
	cancel()
	if err != nil {
		fmt.Println("database unreachable: ", err)
		return 1
	}
	goose.SetBaseFS(migrations.FS)
	err = goose.SetDialect("postgres")
	if err != nil {
		fmt.Println("goose dialect error: ", err)
		return 1
	}
	err = goose.Up(conn, ".")
	if err != nil {
		fmt.Println("goose up error: ", err)
		return 1
	}
	return m.Run()
}

func WithTx(t *testing.T, conn *sql.DB, fn func(ctx context.Context, q *Queries, tx *sql.Tx)) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	tx, err := conn.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		t.Fatalf("begin tx error: %v\n", err)
	}
	t.Cleanup(func() {
		if err := tx.Rollback(); err != nil {
			t.Errorf("tx rollback error: %v\n", err)
		}
	})
	qtx := New(tx)
	fn(ctx, qtx, tx)
}

func insertTestLinks(t *testing.T, ctx context.Context, tx *sql.Tx, inputs ...link) {
	t.Helper()
	for _, input := range inputs {
		_, err := tx.ExecContext(ctx, `INSERT INTO links(original_url, short_name, short_url, created_at)
		VALUES ($1, $2, $3, $4)`, input.OriginalUrl, input.ShortName, input.ShortUrl, input.CreatedAt)
		if err != nil {
			t.Fatalf("insert link error: %v\n", err)
		}
	}
}

func TestGetLinks(t *testing.T) {
	WithTx(t, conn, func(ctx context.Context, q *Queries, tx *sql.Tx) {
		empty, err := q.GetLinks(ctx)
		if err != nil {
			t.Fatalf("unexpected get links error: %v\n", err)
		}
		require.Len(t, empty, 0)
		input1 := link{
			OriginalUrl: "https://example.com/articles/learning-go",
			ShortUrl:    "http://localhost:8080/learn-go",
			ShortName:   "learn-go",
			CreatedAt:   time.Date(2026, time.November, 1, 10, 0, 0, 0, time.UTC),
		}
		input2 := link{
			OriginalUrl: "https://example.org/docs/postgresql",
			ShortUrl:    "http://localhost:8080/pg-docs",
			ShortName:   "pg-docs",
			CreatedAt:   time.Date(2026, time.November, 2, 12, 30, 0, 0, time.UTC),
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO links (original_url, short_name, short_url, created_at)
		VALUES ($1, $2, $3, $4)`, input1.OriginalUrl, input1.ShortName, input1.ShortUrl, input1.CreatedAt)
		if err != nil {
			t.Fatalf("insert link error: %v\n", err)
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO links (original_url, short_name, short_url, created_at)
		VALUES ($1, $2, $3, $4)`, input2.OriginalUrl, input2.ShortName, input2.ShortUrl, input2.CreatedAt)
		if err != nil {
			t.Fatalf("insert link error: %v\n", err)
		}
		links, err := q.GetLinks(ctx)
		if err != nil {
			t.Fatalf("unexpected get links error: %v\n", err)
		}
		require.Len(t, links, 2)
		assert.Equal(t, input1.OriginalUrl, links[1].OriginalUrl)
		assert.Equal(t, input1.ShortName, links[1].ShortName)
		assert.Equal(t, input1.ShortUrl, links[1].ShortUrl)
		assert.Equal(t, input2.OriginalUrl, links[0].OriginalUrl)
		assert.Equal(t, input2.ShortName, links[0].ShortName)
		assert.Equal(t, input2.ShortUrl, links[0].ShortUrl)
	})
}

func TestGetLinksPage(t *testing.T) {
	WithTx(t, conn, func(ctx context.Context, q *Queries, tx *sql.Tx) {
		inputs := []link{
			{
				OriginalUrl: "https://example.org/guides/postgresql",
				ShortName:   "pg-guide",
				ShortUrl:    "http://localhost:8080/pg-guide",
				CreatedAt:   time.Date(2026, time.November, 3, 10, 0, 0, 0, time.UTC),
			},
			{
				OriginalUrl: "https://example.net/docs/gin-routing",
				ShortName:   "gin-routes",
				ShortUrl:    "http://localhost:8080/gin-routes",
				CreatedAt:   time.Date(2026, time.November, 2, 10, 0, 0, 0, time.UTC),
			},
			{
				OriginalUrl: "https://example.com/articles/go-testing",
				ShortName:   "go-tests",
				ShortUrl:    "http://localhost:8080/go-tests",
				CreatedAt:   time.Date(2026, time.November, 1, 10, 0, 0, 0, time.UTC),
			},
		}
		insertTestLinks(t, ctx, tx, inputs...)
		cases := []struct {
			name          string
			offset, limit int32
			want          []link
		}{
			{"[0,2]", 0, 2, []link{inputs[0], inputs[1]}},
			{"[1,2]", 1, 1, []link{inputs[1]}},
			{"[2,4]", 2, 2, []link{inputs[2]}},
			{"[3,5]", 3, 2, []link{}},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				links, err := q.GetLinksPage(ctx, GetLinksPageParams{Offset: c.offset, Limit: c.limit})
				require.NoError(t, err)
				require.Len(t, links, len(c.want))
				for i := range links {
					assert.Equal(t, c.want[i].ShortName, links[i].ShortName)
				}
			})

		}
	})
}

func TestCreateLink(t *testing.T) {
	WithTx(t, conn, func(ctx context.Context, q *Queries, _ *sql.Tx) {
		input := link{
			OriginalUrl: "https://example.com/articles/go-testing",
			ShortName:   "go-tests",
			ShortUrl:    "http://localhost:8080/go-tests",
		}
		link, err := q.CreateLink(ctx, CreateLinkParams{
			OriginalUrl: input.OriginalUrl,
			ShortName:   input.ShortName,
			ShortUrl:    input.ShortUrl,
		})
		if err != nil {
			t.Fatalf("unexpected create link error: %v\n", err)
		}
		assert.Equal(t, input.OriginalUrl, link.OriginalUrl)
		assert.Equal(t, input.ShortName, link.ShortName)
		assert.Equal(t, input.ShortUrl, link.ShortUrl)
	})
}

func TestGetLinkByID(t *testing.T) {
	WithTx(t, conn, func(ctx context.Context, q *Queries, tx *sql.Tx) {
		input := link{
			OriginalUrl: "https://example.org/guides/postgresql",
			ShortName:   "postgres-guide",
			ShortUrl:    "http://localhost:8080/postgres-guide",
		}
		var id int64
		err := tx.QueryRowContext(ctx, `INSERT INTO links (original_url, short_name, short_url)
		VALUES ($1, $2, $3) RETURNING id`, input.OriginalUrl, input.ShortName, input.ShortUrl).Scan(&id)
		if err != nil {
			t.Fatalf("insert link error: %v\n", err)
		}
		link, err := q.GetLinkByID(ctx, id)
		if err != nil {
			t.Fatalf("unexpected get link error: %v\n", err)
		}
		assert.Equal(t, input.OriginalUrl, link.OriginalUrl)
		assert.Equal(t, input.ShortUrl, link.ShortUrl)
		assert.Equal(t, input.ShortName, link.ShortName)
		assert.Equal(t, id, link.ID)
	})
}

func TestUpdateLinkByID(t *testing.T) {
	WithTx(t, conn, func(ctx context.Context, q *Queries, tx *sql.Tx) {
		input := link{
			OriginalUrl: "https://example.com/articles/go-basics",
			ShortName:   "go-basics",
			ShortUrl:    "http://localhost:8080/go-basics",
		}
		var id int64
		err := tx.QueryRowContext(ctx, `INSERT INTO links (original_url, short_name, short_url)
		VALUES ($1, $2, $3) RETURNING id`, input.OriginalUrl, input.ShortName, input.ShortUrl).Scan(&id)
		if err != nil {
			t.Fatalf("insert link error: %v\n", err)
		}
		updateInput := link{
			OriginalUrl: "https://example.com/articles/go-advanced",
			ShortName:   "go-advanced",
			ShortUrl:    "http://localhost:8080/go-advanced",
		}
		link, err := q.UpdateLinkByID(ctx, UpdateLinkByIDParams{
			ID:          id,
			OriginalUrl: updateInput.OriginalUrl,
			ShortName:   updateInput.ShortName,
			ShortUrl:    updateInput.ShortUrl,
		})
		if err != nil {
			t.Fatalf("unexpected update link error: %v\n", err)
		}
		assert.Equal(t, id, link.ID)
		assert.Equal(t, updateInput.OriginalUrl, link.OriginalUrl)
		assert.Equal(t, updateInput.ShortName, link.ShortName)
		assert.Equal(t, updateInput.ShortUrl, link.ShortUrl)
	})
}

func TestDeleteLink(t *testing.T) {
	WithTx(t, conn, func(ctx context.Context, q *Queries, tx *sql.Tx) {
		input := link{
			OriginalUrl: "https://example.org/notes/temporary",
			ShortName:   "temporary-note",
			ShortUrl:    "http://localhost:8080/temporary-note",
		}
		var id int64
		err := tx.QueryRowContext(ctx, `INSERT INTO links (original_url, short_name, short_url)
		VALUES ($1, $2, $3) RETURNING id`, input.OriginalUrl, input.ShortName, input.ShortUrl).Scan(&id)
		if err != nil {
			t.Fatalf("insert link error: %v\n", err)
		}
		_, err = q.DeleteLink(ctx, id)
		if err != nil {
			t.Fatalf("unexpected delete link error: %v\n", err)
		}
		err = tx.QueryRowContext(ctx, `SELECT id FROM links WHERE id = $1`, id).Scan(&id)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		deleted, err := q.DeleteLink(ctx, id)
		if err != nil {
			t.Fatalf("unexpected delete link error: %v\n", err)
		}
		assert.Equal(t, 0, deleted)
	})
}

func TestGetLinkVisitsPage(t *testing.T) {
	WithTx(t, conn, func(ctx context.Context, q *Queries, tx *sql.Tx) {
		linkInput := link{
			OriginalUrl: "https://example.org/guides/testing-redirects",
			ShortName:   "redirect-guide",
			ShortUrl:    "http://localhost:8080/r/redirect-guide",
		}
		var linkId int64
		const status = http.StatusFound
		err := tx.QueryRowContext(ctx, `INSERT INTO links (original_url, short_name, short_url) VALUES ($1, $2, $3) RETURNING id`, linkInput.OriginalUrl, linkInput.ShortName, linkInput.ShortUrl).Scan(&linkId)
		if err != nil {
			t.Fatalf("insert link error: %v\n", err)
		}
		visitInputs := []linkVisit{
			{
				LinkID:    linkId,
				Ip:        "192.0.2.11",
				UserAgent: "curl/8.5.0",
				Referer:   sql.NullString{},
				Status:    status,
				CreatedAt: time.Date(2026, time.November, 20, 10, 0, 0, 0, time.UTC),
			},
			{
				LinkID:    linkId,
				Ip:        "192.0.2.12",
				UserAgent: "Mozilla/5.0",
				Referer:   sql.NullString{String: "https://example.org/articles", Valid: true},
				Status:    status,
				CreatedAt: time.Date(2026, time.November, 20, 11, 0, 0, 0, time.UTC),
			},
			{
				LinkID:    linkId,
				Ip:        "192.0.2.13",
				UserAgent: "Wget/1.21.4",
				Referer:   sql.NullString{String: "https://example.net/resources", Valid: true},
				Status:    status,
				CreatedAt: time.Date(2026, time.November, 20, 12, 0, 0, 0, time.UTC),
			},
		}
		for _, input := range visitInputs {
			_, err := tx.ExecContext(ctx, `INSERT INTO link_visits(link_id, ip, user_agent, referer, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)`, input.LinkID, input.Ip, input.UserAgent, input.Referer, input.Status, input.CreatedAt)
			if err != nil {
				t.Fatalf("insert link visit error: %v\n", err)
			}
		}
		cases := []struct {
			name          string
			offset, limit int32
			want          []linkVisit
		}{
			{"[0,2]", 0, 2, []linkVisit{visitInputs[2], visitInputs[1]}},
			{"[1,2]", 1, 1, []linkVisit{visitInputs[1]}},
			{"[2,4]", 2, 2, []linkVisit{visitInputs[0]}},
			{"[3,5]", 3, 2, []linkVisit{}},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				linkVisits, err := q.GetLinkVisitsPage(ctx, GetLinkVisitsPageParams{Offset: c.offset, Limit: c.limit})
				require.NoError(t, err)
				require.Len(t, linkVisits, len(c.want))
				for i := range linkVisits {
					assert.Equal(t, c.want[i].Ip, linkVisits[i].Ip)
				}
			})
		}
	})
}
