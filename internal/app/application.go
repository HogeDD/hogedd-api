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
	accessTokenVerifier httpapi.AccessTokenVerifier
	userGetter          userhttp.CurrentUserGetter
	userRegistrar       userhttp.AuthenticatedUserRegistrar
}

// WithUserGetter は認証済みUser取得endpointが使用するUse Caseを設定します。
func WithUserGetter(getter userhttp.CurrentUserGetter) Option {
	return func(dependencies *dependencies) {
		if getter != nil {
			dependencies.userGetter = getter
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

func (unavailableUserGetter) Execute(
	context.Context,
	identity.Identity,
) (userapp.RegisteredUserResult, error) {
	return userapp.RegisteredUserResult{}, userapp.ErrRegistrationFailed
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
	handler                 http.Handler
	healthHandler           http.Handler
	appsListHandler         http.Handler
	appDetailHandler        http.Handler
	meHandler               http.Handler
	userRegistrationHandler http.Handler
}

// New はロガーを受け取り、実行可能なApplicationを構築します。
// loggerがnilの場合はslogのデフォルトロガーを使用します。
func New(logger *slog.Logger, options ...Option) (*Application, error) {
	if logger == nil {
		logger = slog.Default()
	}
	dependencies := dependencies{
		accessTokenVerifier: rejectAccessTokenVerifier{},
		userGetter:          unavailableUserGetter{},
		userRegistrar:       unavailableUserRegistrar{},
	}
	for _, option := range options {
		option(&dependencies)
	}

	responder := httpapi.NewResponder(logger)
	healthService := health.NewService()
	healthHandler := httpapi.NewHealthHandler(healthService, responder)
	meHandler := identityhttp.NewMeHandler(responder)
	userMeHandler := userhttp.NewMeHandler(dependencies.userGetter, dependencies.userRegistrar, responder)
	appStore, err := infrastructure.NewSeededMemoryAppStore()
	if err != nil {
		return nil, err
	}
	appsHandler := contenthttp.NewAppsHandler(
		application.NewListPublishedAppsUseCase(appStore),
		application.NewGetPublishedAppUseCase(appStore),
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

	return &Application{
		handler: middleware.Wrap(httpapi.NewRouter(
			healthHandler,
			http.HandlerFunc(appsHandler.List),
			http.HandlerFunc(appsHandler.Get),
			authenticatedMeHandler,
			authenticatedUserMeHandler,
		)),
		healthHandler:           middleware.Wrap(healthHandler),
		appsListHandler:         middleware.Wrap(http.HandlerFunc(appsHandler.List)),
		appDetailHandler:        middleware.Wrap(http.HandlerFunc(appsHandler.Get)),
		meHandler:               middleware.Wrap(authenticatedMeHandler),
		userRegistrationHandler: middleware.Wrap(authenticatedUserMeHandler),
	}, nil
}

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

// AppDetailHandler はVercelの公開アプリ詳細Functionで使用するHandlerを返します。
func (a *Application) AppDetailHandler() http.Handler {
	return a.appDetailHandler
}

// MeHandler はVercelの認証主体取得Functionで使用するHandlerを返します。
func (a *Application) MeHandler() http.Handler {
	return a.meHandler
}
