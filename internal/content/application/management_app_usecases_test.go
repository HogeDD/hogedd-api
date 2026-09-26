package application

import (
	"context"
	"testing"
	"time"

	"github.com/iwasawa/hogedd-api/internal/content/domain"
)

type managementAppStoreStub struct {
	apps    []*domain.App
	created *domain.App
	err     error
	version int64
	found   bool
}

func (s *managementAppStoreStub) FindForManagement(context.Context, domain.Slug) (*domain.App, int64, bool, error) {
	if len(s.apps) == 0 {
		return nil, 0, false, s.err
	}
	return s.apps[0], s.version, s.found, s.err
}

func (s *managementAppStoreStub) UpdateDraft(_ context.Context, app *domain.App, expectedVersion int64) (int64, error) {
	if s.err != nil {
		return 0, s.err
	}
	s.created = app
	return expectedVersion + 1, nil
}

func (s *managementAppStoreStub) Update(ctx context.Context, app *domain.App, expectedVersion int64) (int64, error) {
	return s.UpdateDraft(ctx, app, expectedVersion)
}

func (s *managementAppStoreStub) Publish(_ context.Context, app *domain.App, expectedVersion int64) (int64, error) {
	if s.err != nil {
		return 0, s.err
	}
	s.created = app
	return expectedVersion + 1, nil
}

func (s *managementAppStoreStub) List(context.Context) ([]*domain.App, error) { return s.apps, s.err }
func (s *managementAppStoreStub) Create(_ context.Context, app *domain.App) error {
	s.created = app
	return s.err
}

func TestCreatePreparingAppUseCase(t *testing.T) {
	store := &managementAppStoreStub{}
	result, err := NewCreatePreparingAppUseCase(store).Execute(context.Background(), "new-app", " New App ", " Description ", []string{"Game"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Slug != "new-app" || result.Status != "preparing" || store.created == nil {
		t.Fatalf("result/store = %+v/%+v", result, store.created)
	}
}

func TestListManagementAppsIncludesPreparing(t *testing.T) {
	slug, _ := domain.NewSlug("draft-app")
	app, _ := domain.NewPreparingApp(slug, "Draft", "Description", nil)
	result, err := NewListManagementAppsUseCase(&managementAppStoreStub{apps: []*domain.App{app}}).Execute(context.Background())
	if err != nil || len(result) != 1 || result[0].Status != "preparing" {
		t.Fatalf("Execute() = %+v, %v", result, err)
	}
}

func TestGetAndUpdateManagementApp(t *testing.T) {
	slug, _ := domain.NewSlug("draft-app")
	draft, _ := domain.NewPreparingApp(slug, "Draft", "Description", nil)
	store := &managementAppStoreStub{apps: []*domain.App{draft}, version: 3, found: true}
	detail, err := NewGetManagementAppUseCase(store).Execute(context.Background(), "draft-app")
	if err != nil || detail.Version != 3 {
		t.Fatalf("Get Execute() = %+v, %v", detail, err)
	}
	updated, err := NewUpdateManagementAppUseCase(store).Execute(context.Background(), "draft-app", " Updated ", " New description ", []string{"Go"}, "private", "", "", detail.Version)
	if err != nil || updated.Title != "Updated" || updated.Version != 4 {
		t.Fatalf("Update Execute() = %+v, %v", updated, err)
	}
}

func TestUpdateManagementAppRejectsInvalidVersion(t *testing.T) {
	_, err := NewUpdateManagementAppUseCase(&managementAppStoreStub{}).Execute(context.Background(), "draft-app", "Draft", "Description", nil, "private", "", "", 0)
	if err != ErrAppVersionConflict {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestPublishManagementApp(t *testing.T) {
	slug, _ := domain.NewSlug("draft-app")
	draft, _ := domain.NewPreparingApp(slug, "Draft", "Description", nil)
	store := &managementAppStoreStub{apps: []*domain.App{draft}, version: 2, found: true}
	now := time.Date(2026, time.September, 26, 10, 0, 0, 0, time.UTC)
	result, err := NewPublishManagementAppUseCase(store, func() time.Time { return now }).Execute(context.Background(), "draft-app", "学習DD", "https://youtu.be/video", 2)
	if err != nil || result.Status != "published" || result.Version != 3 || result.PublishedAt != now.Format(time.RFC3339) {
		t.Fatalf("Execute() = %+v, %v", result, err)
	}
	store.version = 3
	result, err = NewPublishManagementAppUseCase(store, func() time.Time { return now.Add(time.Hour) }).Execute(context.Background(), "draft-app", "ignored", "https://youtu.be/other", 3)
	if err != nil || result.Version != 3 || result.PublishedAt != now.Format(time.RFC3339) {
		t.Fatalf("idempotent Execute() = %+v, %v", result, err)
	}
}
