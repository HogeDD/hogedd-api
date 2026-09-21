package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/iwasawa/hogedd-api/internal/content/domain"
)

var errTestStore = errors.New("test store error")

type stubPublishedApps struct {
	apps      []*domain.App
	listErr   error
	found     bool
	findErr   error
	foundSlug domain.Slug
}

func (s *stubPublishedApps) List(context.Context) ([]*domain.App, error) {
	return s.apps, s.listErr
}

func (s *stubPublishedApps) FindBySlug(_ context.Context, slug domain.Slug) (*domain.App, bool, error) {
	s.foundSlug = slug
	if s.findErr != nil || !s.found {
		return nil, s.found, s.findErr
	}
	return s.apps[0], true, nil
}

func TestListPublishedAppsExecuteOmitsPreparingApps(t *testing.T) {
	t.Parallel()

	preparing := mustPreparingApp(t, "preparing-app")
	useCase := NewListPublishedAppsUseCase(&stubPublishedApps{
		apps: []*domain.App{mustPublishedApp(t, "clean-tasks"), preparing},
	})

	got, err := useCase.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if len(got) != 1 || got[0].Slug != "clean-tasks" {
		t.Fatalf("Execute() = %+v, want only published app", got)
	}
}

func TestListPublishedAppsExecute(t *testing.T) {
	t.Parallel()

	app := mustPublishedApp(t, "clean-tasks")
	useCase := NewListPublishedAppsUseCase(&stubPublishedApps{apps: []*domain.App{app}})

	got, err := useCase.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("Execute() returned %d apps, want 1", len(got))
	}
	assertPublishedApp(t, got[0], "clean-tasks")
}

func TestListPublishedAppsExecutePropagatesPortError(t *testing.T) {
	t.Parallel()

	useCase := NewListPublishedAppsUseCase(&stubPublishedApps{listErr: errTestStore})
	_, err := useCase.Execute(context.Background())
	if !errors.Is(err, errTestStore) {
		t.Fatalf("Execute() error = %v, want %v", err, errTestStore)
	}
}

func TestGetPublishedAppExecute(t *testing.T) {
	t.Parallel()

	store := &stubPublishedApps{
		apps:  []*domain.App{mustPublishedApp(t, "clean-tasks")},
		found: true,
	}
	useCase := NewGetPublishedAppUseCase(store)

	got, err := useCase.Execute(context.Background(), "clean-tasks")
	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	assertPublishedApp(t, got, "clean-tasks")
	if store.foundSlug.String() != "clean-tasks" {
		t.Fatalf("FindBySlug() slug = %q, want %q", store.foundSlug.String(), "clean-tasks")
	}
}

func TestGetPublishedAppExecuteHidesPreparingApp(t *testing.T) {
	t.Parallel()

	useCase := NewGetPublishedAppUseCase(&stubPublishedApps{
		apps:  []*domain.App{mustPreparingApp(t, "preparing-app")},
		found: true,
	})

	_, err := useCase.Execute(context.Background(), "preparing-app")
	if !errors.Is(err, ErrAppNotFound) {
		t.Fatalf("Execute() error = %v, want %v", err, ErrAppNotFound)
	}
}

func TestGetPublishedAppExecuteReturnsNotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		slug  string
		store *stubPublishedApps
	}{
		{name: "存在しない", slug: "unknown", store: &stubPublishedApps{}},
		{name: "不正なslug", slug: "Invalid/Slug", store: &stubPublishedApps{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			useCase := NewGetPublishedAppUseCase(tt.store)
			_, err := useCase.Execute(context.Background(), tt.slug)
			if !errors.Is(err, ErrAppNotFound) {
				t.Fatalf("Execute() error = %v, want %v", err, ErrAppNotFound)
			}
		})
	}
}

func TestGetPublishedAppExecutePropagatesPortError(t *testing.T) {
	t.Parallel()

	useCase := NewGetPublishedAppUseCase(&stubPublishedApps{findErr: errTestStore})
	_, err := useCase.Execute(context.Background(), "clean-tasks")
	if !errors.Is(err, errTestStore) {
		t.Fatalf("Execute() error = %v, want %v", err, errTestStore)
	}
}

func mustPublishedApp(t *testing.T, slugValue string) *domain.App {
	t.Helper()

	app := mustPreparingApp(t, slugValue)
	if err := app.Publish(
		time.Date(2026, time.May, 30, 0, 0, 0, 0, time.UTC),
		"学習DD",
		"https://youtu.be/5mo0qnPuVTY",
	); err != nil {
		t.Fatalf("Publish() unexpected error = %v", err)
	}
	return app
}

func mustPreparingApp(t *testing.T, slugValue string) *domain.App {
	t.Helper()

	slug, err := domain.NewSlug(slugValue)
	if err != nil {
		t.Fatalf("NewSlug() unexpected error = %v", err)
	}
	app, err := domain.NewPreparingApp(slug, "Clean Tasks", "タスク管理アプリ。", []string{"Next.js"})
	if err != nil {
		t.Fatalf("NewPreparingApp() unexpected error = %v", err)
	}
	return app
}

func assertPublishedApp(t *testing.T, got PublishedAppResult, wantSlug string) {
	t.Helper()

	if got.Slug != wantSlug {
		t.Fatalf("Slug = %q, want %q", got.Slug, wantSlug)
	}
	if got.Title != "Clean Tasks" {
		t.Fatalf("Title = %q, want %q", got.Title, "Clean Tasks")
	}
	if got.PublishedAt.IsZero() {
		t.Fatal("PublishedAt is zero")
	}
	if got.YouTubeURL == "" {
		t.Fatal("YouTubeURL is empty")
	}
}
