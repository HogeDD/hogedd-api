package application

import (
	"context"
	"testing"

	"github.com/iwasawa/hogedd-api/internal/content/domain"
)

type managementAppStoreStub struct {
	apps    []*domain.App
	created *domain.App
	err     error
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
