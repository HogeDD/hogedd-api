package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	contentapp "github.com/iwasawa/hogedd-api/internal/content/application"
	"github.com/iwasawa/hogedd-api/internal/content/domain"
	contentinfra "github.com/iwasawa/hogedd-api/internal/content/infrastructure"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPostgresAppRepositoryCreateAndList(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	const slugValue = "repository-integration-test"
	if _, err := database.ExecContext(ctx, "DELETE FROM content_apps WHERE slug = $1", slugValue); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = database.ExecContext(context.Background(), "DELETE FROM content_apps WHERE slug = $1", slugValue)
		_ = database.Close()
	})

	slug, _ := domain.NewSlug(slugValue)
	app, _ := domain.NewPreparingApp(slug, "Repository Test", "PostgreSQL adapter test", []string{"Go", "PostgreSQL"})
	repository := contentinfra.NewPostgresAppRepository(database)
	if err := repository.Create(ctx, app); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := repository.Create(ctx, app); !errors.Is(err, contentapp.ErrAppAlreadyExists) {
		t.Fatalf("duplicate Create() error = %v", err)
	}
	stored, version, found, err := repository.FindForManagement(ctx, slug)
	if err != nil || !found || version != 1 {
		t.Fatalf("FindForManagement() = %+v, %d, %v, %v", stored, version, found, err)
	}
	if err := stored.UpdateDraftDetails("Updated Repository Test", "Updated adapter test", []string{"Go"}); err != nil {
		t.Fatal(err)
	}
	newVersion, err := repository.UpdateDraft(ctx, stored, version)
	if err != nil || newVersion != 2 {
		t.Fatalf("UpdateDraft() = %d, %v", newVersion, err)
	}
	if _, err := repository.UpdateDraft(ctx, stored, version); !errors.Is(err, contentapp.ErrAppVersionConflict) {
		t.Fatalf("stale UpdateDraft() error = %v", err)
	}
	apps, err := repository.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	for _, stored := range apps {
		if stored.Slug().String() == slugValue && stored.Title() == "Updated Repository Test" {
			return
		}
	}
	t.Fatalf("List() did not return %q", slugValue)
}
