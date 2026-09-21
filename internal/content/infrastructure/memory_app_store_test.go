package infrastructure

import (
	"context"
	"errors"
	"testing"

	"github.com/iwasawa/hogedd-api/internal/content/application"
	"github.com/iwasawa/hogedd-api/internal/content/domain"
)

func TestSeededMemoryAppStoreWithUseCases(t *testing.T) {
	t.Parallel()

	store, err := NewSeededMemoryAppStore()
	if err != nil {
		t.Fatalf("NewSeededMemoryAppStore() error = %v", err)
	}

	all, err := store.List(context.Background())
	if err != nil || len(all) != 5 {
		t.Fatalf("List() = %d apps, error %v; want 5 apps", len(all), err)
	}

	list := application.NewListPublishedAppsUseCase(store)
	published, err := list.Execute(context.Background())
	if err != nil {
		t.Fatalf("ListPublishedAppsUseCase.Execute() error = %v", err)
	}
	if len(published) != 1 || published[0].Slug != "clean-tasks" {
		t.Fatalf("published apps = %+v, want only clean-tasks", published)
	}

	get := application.NewGetPublishedAppUseCase(store)
	_, err = get.Execute(context.Background(), "chinchin")
	if !errors.Is(err, application.ErrAppNotFound) {
		t.Fatalf("GetPublishedAppUseCase.Execute(chinchin) error = %v, want not found", err)
	}
	got, err := get.Execute(context.Background(), "clean-tasks")
	if err != nil || got.Title != "Clean Tasks" {
		t.Fatalf("GetPublishedAppUseCase.Execute(clean-tasks) = %+v, %v", got, err)
	}
}

func TestMemoryAppStoreCancellation(t *testing.T) {
	t.Parallel()

	store := NewMemoryAppStore(nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := store.List(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("List() error = %v, want context canceled", err)
	}
	slug, _ := domain.NewSlug("clean-tasks")
	if _, _, err := store.FindBySlug(ctx, slug); !errors.Is(err, context.Canceled) {
		t.Fatalf("FindBySlug() error = %v, want context canceled", err)
	}
}
