package shared

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/iwasawa/hogedd-api/internal/app"
	"github.com/iwasawa/hogedd-api/internal/config"
	contentapp "github.com/iwasawa/hogedd-api/internal/content/application"
	contentinfra "github.com/iwasawa/hogedd-api/internal/content/infrastructure"
	platformpostgres "github.com/iwasawa/hogedd-api/internal/platform/postgres"
)

var application = sync.OnceValue(newApplication)

// Application は公開Apps Functionで共有するDB接続済みApplicationを返します。
func Application() *app.Application { return application() }

func newApplication() *app.Application {
	runtimeConfig := config.LoadRuntime()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: runtimeConfig.LogLevel}))
	databaseConfig, err := config.LoadDatabase()
	if err != nil {
		panic(err)
	}
	database, err := platformpostgres.Open(context.Background(), databaseConfig)
	if err != nil {
		panic(err)
	}
	apps := contentinfra.NewPostgresAppRepository(database)
	application, err := app.New(logger,
		app.WithPublishedAppUseCases(contentapp.NewListPublishedAppsUseCase(apps), contentapp.NewGetPublishedAppUseCase(apps), contentapp.NewListRecommendedAppsUseCase(apps)),
		app.WithAppMetrics(contentapp.NewRecordAppLaunchUseCase(apps, nil), contentapp.NewGetAppRecommendationsUseCase(apps, nil), config.LoadMetricsIngestToken()),
	)
	if err != nil {
		panic(err)
	}
	return application
}
