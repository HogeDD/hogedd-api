package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/iwasawa/hogedd-api/internal/content/application"
	"github.com/iwasawa/hogedd-api/internal/content/infrastructure"
	contenthttp "github.com/iwasawa/hogedd-api/internal/content/transport/http"
	"github.com/iwasawa/hogedd-api/internal/health"
	"github.com/iwasawa/hogedd-api/internal/identity"
	identityhttp "github.com/iwasawa/hogedd-api/internal/identity/transport/http"
	"github.com/iwasawa/hogedd-api/internal/transport/httpapi"
	userapp "github.com/iwasawa/hogedd-api/internal/user/application"
	userhttp "github.com/iwasawa/hogedd-api/internal/user/transport/http"
)

type dependencies struct {
	publishedAppsLister    contenthttp.PublishedAppsLister
	publishedAppGetter     contenthttp.PublishedAppGetter
	recommendedAppsLister  contenthttp.PublishedAppsLister
	accessTokenVerifier    httpapi.AccessTokenVerifier
	userGetter             userhttp.CurrentUserGetter
	userRegistrar          userhttp.AuthenticatedUserRegistrar
	profileGetter          userhttp.CurrentProfileGetter
	profileUpdater         userhttp.CurrentProfileUpdater
	managementUserGetter   userhttp.ManagementUserGetter
	managementAppsLister   contenthttp.ManagementAppsLister
	preparingAppCreator    contenthttp.PreparingAppCreator
	managementAppGetter    contenthttp.ManagementAppGetter
	managementAppUpdater   contenthttp.ManagementAppUpdater
	managementAppPublisher contenthttp.ManagementAppPublisher
}

// WithPublishedAppUseCases は公開Appの一覧・詳細・おすすめ取得元を設定します。
func WithPublishedAppUseCases(list contenthttp.PublishedAppsLister, get contenthttp.PublishedAppGetter, recommended contenthttp.PublishedAppsLister) Option {
	return func(dependencies *dependencies) {
		if list != nil {
			dependencies.publishedAppsLister = list
		}
		if get != nil {
			dependencies.publishedAppGetter = get
		}
		if recommended != nil {
			dependencies.recommendedAppsLister = recommended
		}
	}
}

// WithManagementAppPublicationUseCase は運営App公開Use Caseを設定します。
func WithManagementAppPublicationUseCase(publisher contenthttp.ManagementAppPublisher) Option {
	return func(dependencies *dependencies) {
		if publisher != nil {
			dependencies.managementAppPublisher = publisher
		}
	}
}

// WithManagementAppDetailUseCases は運営App詳細の取得・更新Use Caseを設定します。
func WithManagementAppDetailUseCases(getter contenthttp.ManagementAppGetter, updater contenthttp.ManagementAppUpdater) Option {
	return func(dependencies *dependencies) {
		if getter != nil {
			dependencies.managementAppGetter = getter
		}
		if updater != nil {
			dependencies.managementAppUpdater = updater
		}
	}
}

// WithManagementAppUseCases は運営App一覧・作成Use Caseを設定します。
func WithManagementAppUseCases(lister contenthttp.ManagementAppsLister, creator contenthttp.PreparingAppCreator) Option {
	return func(dependencies *dependencies) {
		if lister != nil {
			dependencies.managementAppsLister = lister
		}
		if creator != nil {
			dependencies.preparingAppCreator = creator
		}
	}
}

// WithProfileUseCases は現在Userのプロフィール取得・更新Use Caseを設定します。
func WithProfileUseCases(getter userhttp.CurrentProfileGetter, updater userhttp.CurrentProfileUpdater) Option {
	return func(dependencies *dependencies) {
		if getter != nil {
			dependencies.profileGetter = getter
		}
		if updater != nil {
			dependencies.profileUpdater = updater
		}
	}
}

// WithUserGetter は認証済みUser取得endpointが使用するUse Caseを設定します。
func WithUserGetter(getter userhttp.CurrentUserGetter) Option {
	return func(dependencies *dependencies) {
		if getter != nil {
			dependencies.userGetter = getter
		}
	}
}

// WithManagementUserGetter は運営境界が使用する認可Use Caseを設定します。
func WithManagementUserGetter(getter userhttp.ManagementUserGetter) Option {
	return func(dependencies *dependencies) {
		if getter != nil {
			dependencies.managementUserGetter = getter
		}
	}
}

// WithUserRegistrar は認証済みUser登録endpointが使用するUse Caseを設定します。
func WithUserRegistrar(registrar userhttp.AuthenticatedUserRegistrar) Option {
	return func(dependencies *dependencies) {
		if registrar != nil {
			dependencies.userRegistrar = registrar
		}
	}
}

// Option はApplicationが使用する外部依存を差し替えます。
type Option func(*dependencies)

// WithAccessTokenVerifier は保護endpointが使用するAccess Token検証器を設定します。
func WithAccessTokenVerifier(verifier httpapi.AccessTokenVerifier) Option {
	return func(dependencies *dependencies) {
		if verifier != nil {
			dependencies.accessTokenVerifier = verifier
		}
	}
}

type rejectAccessTokenVerifier struct{}

func (rejectAccessTokenVerifier) Verify(context.Context, string) (identity.Identity, error) {
	return identity.Identity{}, errors.New("access token verifier is not configured")
}

type unavailableUserRegistrar struct{}

type unavailableUserGetter struct{}

type unavailableManagementUserGetter struct{}

type unavailableProfileUseCase struct{}
type unavailableManagementApps struct{}

func (unavailableManagementApps) Execute(context.Context) ([]application.ManagementAppResult, error) {
	return nil, errors.New("management apps are not configured")
}
func (unavailableManagementApps) ExecuteCreate(context.Context, string, string, string, []string) (application.ManagementAppResult, error) {
	return application.ManagementAppResult{}, errors.New("management apps are not configured")
}

type unavailablePreparingAppCreator struct{ unavailableManagementApps }

type unavailableManagementAppGetter struct{ unavailableManagementApps }
type unavailableManagementAppUpdater struct{ unavailableManagementApps }
type unavailableManagementAppPublisher struct{ unavailableManagementApps }

func (unavailableManagementAppGetter) Execute(context.Context, string) (application.ManagementAppResult, error) {
	return application.ManagementAppResult{}, errors.New("management app is not configured")
}

func (unavailableManagementAppUpdater) Execute(context.Context, string, string, string, []string, int64) (application.ManagementAppResult, error) {
	return application.ManagementAppResult{}, errors.New("management app is not configured")
}

func (unavailableManagementAppPublisher) Execute(context.Context, string, string, string, int64) (application.ManagementAppResult, error) {
	return application.ManagementAppResult{}, errors.New("management app publication is not configured")
}

func (unavailablePreparingAppCreator) Execute(ctx context.Context, slug, title, description string, tags []string) (application.ManagementAppResult, error) {
	return unavailableManagementApps{}.ExecuteCreate(ctx, slug, title, description, tags)
}

func (unavailableProfileUseCase) Execute(context.Context, identity.Identity) (userapp.ProfileResult, error) {
	return userapp.ProfileResult{}, userapp.ErrUserNotFound
}

func (unavailableProfileUseCase) ExecuteUpdate(context.Context, identity.Identity, string) (userapp.ProfileResult, error) {
	return userapp.ProfileResult{}, userapp.ErrUserNotFound
}

type unavailableProfileUpdater struct{ unavailableProfileUseCase }

func (unavailableProfileUpdater) Execute(ctx context.Context, authenticated identity.Identity, displayName string) (userapp.ProfileResult, error) {
	return unavailableProfileUseCase{}.ExecuteUpdate(ctx, authenticated, displayName)
}

func (unavailableUserGetter) Execute(
	context.Context,
	identity.Identity,
) (userapp.RegisteredUserResult, error) {
	return userapp.RegisteredUserResult{}, userapp.ErrRegistrationFailed
}

func (unavailableManagementUserGetter) Execute(
	context.Context,
	identity.Identity,
) (userapp.ManagementUserResult, error) {
	return userapp.ManagementUserResult{}, userapp.ErrManagementUserNotFound
}

func (unavailableUserRegistrar) Execute(
	context.Context,
	identity.Identity,
	string,
) (userapp.RegisteredUserResult, error) {
	return userapp.RegisteredUserResult{}, userapp.ErrRegistrationFailed
}

// Application はアプリケーション全体の依存関係を保持するコンポジションルートです。
// 具体的な実装の組み立てをこの型へ集約し、各機能から依存生成の責務を分離します。
type Application struct {
	handler                         http.Handler
	healthHandler                   http.Handler
	appsListHandler                 http.Handler
	appsRecommendedHandler          http.Handler
	appDetailHandler                http.Handler
	meHandler                       http.Handler
	userRegistrationHandler         http.Handler
	userProfileHandler              http.Handler
	managementUserHandler           http.Handler
	managementAppsHandler           http.Handler
	managementAppHandler            http.Handler
	managementAppPublicationHandler http.Handler
	managementNotFoundHandler       http.Handler
}

// New はロガーを受け取り、実行可能なApplicationを構築します。
// loggerがnilの場合はslogのデフォルトロガーを使用します。
func New(logger *slog.Logger, options ...Option) (*Application, error) {
	if logger == nil {
		logger = slog.Default()
	}
	dependencies := dependencies{
		publishedAppsLister: nil, publishedAppGetter: nil, recommendedAppsLister: nil,
		accessTokenVerifier:    rejectAccessTokenVerifier{},
		userGetter:             unavailableUserGetter{},
		userRegistrar:          unavailableUserRegistrar{},
		profileGetter:          unavailableProfileUseCase{},
		profileUpdater:         unavailableProfileUpdater{},
		managementUserGetter:   unavailableManagementUserGetter{},
		managementAppsLister:   unavailableManagementApps{},
		preparingAppCreator:    unavailablePreparingAppCreator{},
		managementAppGetter:    unavailableManagementAppGetter{},
		managementAppUpdater:   unavailableManagementAppUpdater{},
		managementAppPublisher: unavailableManagementAppPublisher{},
	}
	for _, option := range options {
		option(&dependencies)
	}

	responder := httpapi.NewResponder(logger)
	healthService := health.NewService()
	healthHandler := httpapi.NewHealthHandler(healthService, responder)
	meHandler := identityhttp.NewMeHandler(responder)
	userMeHandler := userhttp.NewMeHandler(dependencies.userGetter, dependencies.userRegistrar, responder)
	userProfileHandler := userhttp.NewProfileHandler(dependencies.profileGetter, dependencies.profileUpdater, responder)
	managementUserHandler := userhttp.NewManagementHandler(dependencies.managementUserGetter, responder)
	managementAppsHandler := contenthttp.NewManagementAppsHandler(dependencies.managementUserGetter, dependencies.managementAppsLister, dependencies.preparingAppCreator, responder)
	managementAppHandler := contenthttp.NewManagementAppHandler(dependencies.managementUserGetter, dependencies.managementAppGetter, dependencies.managementAppUpdater, responder)
	managementAppPublicationHandler := contenthttp.NewManagementAppPublicationHandler(dependencies.managementUserGetter, dependencies.managementAppPublisher, responder)
	managementNotFoundHandler := responder.NotFoundHandler()
	appStore, err := infrastructure.NewSeededMemoryAppStore()
	if err != nil {
		return nil, err
	}
	if dependencies.publishedAppsLister == nil {
		dependencies.publishedAppsLister = application.NewListPublishedAppsUseCase(appStore)
	}
	if dependencies.publishedAppGetter == nil {
		dependencies.publishedAppGetter = application.NewGetPublishedAppUseCase(appStore)
	}
	if dependencies.recommendedAppsLister == nil {
		dependencies.recommendedAppsLister = application.NewListPublishedAppsUseCase(appStore)
	}
	appsHandler := contenthttp.NewAppsHandler(
		dependencies.publishedAppsLister, dependencies.publishedAppGetter, dependencies.recommendedAppsLister,
		responder,
	)
	middleware := httpapi.NewMiddlewareStack(
		httpapi.RequestID(),
		httpapi.SecurityHeaders(),
		httpapi.AccessLog(logger),
		httpapi.Recover(logger, responder),
	)
	authenticatedMeHandler := httpapi.Chain(
		meHandler,
		httpapi.AuthenticateBearer(dependencies.accessTokenVerifier, responder),
	)
	authenticatedUserMeHandler := httpapi.Chain(
		userMeHandler,
		httpapi.AuthenticateBearer(dependencies.accessTokenVerifier, responder),
	)
	authenticatedUserProfileHandler := httpapi.Chain(
		userProfileHandler,
		httpapi.AuthenticateBearer(dependencies.accessTokenVerifier, responder),
	)
	concealedManagementUserHandler := httpapi.Chain(
		managementUserHandler,
		httpapi.ConcealBearer(dependencies.accessTokenVerifier, responder),
	)
	concealedManagementAppsHandler := httpapi.Chain(managementAppsHandler, httpapi.ConcealBearer(dependencies.accessTokenVerifier, responder))
	concealedManagementAppHandler := httpapi.Chain(managementAppHandler, httpapi.ConcealBearer(dependencies.accessTokenVerifier, responder))
	concealedManagementAppPublicationHandler := httpapi.Chain(managementAppPublicationHandler, httpapi.ConcealBearer(dependencies.accessTokenVerifier, responder))

	return &Application{
		handler: middleware.Wrap(httpapi.NewRouter(
			healthHandler,
			http.HandlerFunc(appsHandler.List),
			http.HandlerFunc(appsHandler.Recommended),
			http.HandlerFunc(appsHandler.Get),
			authenticatedMeHandler,
			authenticatedUserMeHandler,
			authenticatedUserProfileHandler,
			concealedManagementUserHandler,
			concealedManagementAppsHandler,
			concealedManagementAppHandler,
			concealedManagementAppPublicationHandler,
			managementNotFoundHandler,
		)),
		healthHandler:                   middleware.Wrap(healthHandler),
		appsListHandler:                 middleware.Wrap(http.HandlerFunc(appsHandler.List)),
		appsRecommendedHandler:          middleware.Wrap(http.HandlerFunc(appsHandler.Recommended)),
		appDetailHandler:                middleware.Wrap(http.HandlerFunc(appsHandler.Get)),
		meHandler:                       middleware.Wrap(authenticatedMeHandler),
		userRegistrationHandler:         middleware.Wrap(authenticatedUserMeHandler),
		userProfileHandler:              middleware.Wrap(authenticatedUserProfileHandler),
		managementUserHandler:           middleware.Wrap(concealedManagementUserHandler),
		managementAppsHandler:           middleware.Wrap(concealedManagementAppsHandler),
		managementAppHandler:            middleware.Wrap(concealedManagementAppHandler),
		managementAppPublicationHandler: middleware.Wrap(concealedManagementAppPublicationHandler),
		managementNotFoundHandler:       middleware.Wrap(managementNotFoundHandler),
	}, nil
}

// ManagementAppsHandler はVercelの運営App一覧・作成Functionで使用するHandlerを返します。
func (a *Application) ManagementAppsHandler() http.Handler { return a.managementAppsHandler }

// ManagementAppHandler はVercelの運営App詳細Functionで使用するHandlerを返します。
func (a *Application) ManagementAppHandler() http.Handler { return a.managementAppHandler }

// ManagementAppPublicationHandler はVercelの運営App公開Functionで使用するHandlerを返します。
func (a *Application) ManagementAppPublicationHandler() http.Handler {
	return a.managementAppPublicationHandler
}

// ManagementUserHandler はVercelの運営User確認Functionで使用するHandlerを返します。
func (a *Application) ManagementUserHandler() http.Handler { return a.managementUserHandler }

// ManagementNotFoundHandler は未知の運営routeへ秘匿404を返します。
func (a *Application) ManagementNotFoundHandler() http.Handler { return a.managementNotFoundHandler }

// UserProfileHandler はVercelの現在UserプロフィールFunctionで使用するHandlerを返します。
func (a *Application) UserProfileHandler() http.Handler { return a.userProfileHandler }

// UserRegistrationHandler はVercelの認証済みUser登録Functionで使用するHandlerを返します。
func (a *Application) UserRegistrationHandler() http.Handler {
	return a.userRegistrationHandler
}

// Handler はローカルサーバで全ルートを提供するHTTPハンドラーを返します。
func (a *Application) Handler() http.Handler {
	return a.handler
}

// HealthHandler はVercelのヘルスチェックFunctionで使用するHTTPハンドラーを返します。
func (a *Application) HealthHandler() http.Handler {
	return a.healthHandler
}

// AppsListHandler はVercelの公開アプリ一覧Functionで使用するHandlerを返します。
func (a *Application) AppsListHandler() http.Handler {
	return a.appsListHandler
}

// AppsRecommendedHandler はVercelの公開おすすめアプリFunctionで使用するHandlerを返します。
func (a *Application) AppsRecommendedHandler() http.Handler {
	return a.appsRecommendedHandler
}

// AppDetailHandler はVercelの公開アプリ詳細Functionで使用するHandlerを返します。
func (a *Application) AppDetailHandler() http.Handler {
	return a.appDetailHandler
}

// MeHandler はVercelの認証主体取得Functionで使用するHandlerを返します。
func (a *Application) MeHandler() http.Handler {
	return a.meHandler
}
