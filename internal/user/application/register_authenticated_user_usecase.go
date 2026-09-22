package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iwasawa/hogedd-api/internal/identity"
	userdomain "github.com/iwasawa/hogedd-api/internal/user/domain"
)

var (
	// ErrAccessTokenRequired はAuth0 user profile取得に必要なAccess Tokenがないことを表します。
	ErrAccessTokenRequired = errors.New("access token is required")
	// ErrProfileIdentityMismatch はJWTとAuth0 user profileが異なる認証主体を表すことを示します。
	ErrProfileIdentityMismatch = errors.New("authenticated identity and user profile do not match")
	// ErrProfileUnavailable はAuth0から認証主体のprofileを取得できないことを表します。
	ErrProfileUnavailable = errors.New("authenticated profile is unavailable")
	// ErrRegistrationFailed はUserを永続化できないことを表します。
	ErrRegistrationFailed = errors.New("user registration failed")
)

// AuthenticatedProfile はAuth0がAccess Tokenの主体について返す連絡先snapshotです。
type AuthenticatedProfile struct {
	Subject       string
	Email         string
	EmailVerified bool
}

// AuthenticatedProfileProvider はAccess Tokenの主体に紐づくprofileを取得するportです。
type AuthenticatedProfileProvider interface {
	Fetch(context.Context, string) (AuthenticatedProfile, error)
}

// RegisterUserParams は認証主体からHogeDD Userを登録するための値です。
type RegisterUserParams struct {
	AuthIssuer    string
	AuthSubject   string
	Email         string
	EmailVerified bool
	InitialRole   userdomain.Role
}

// UserRegistrar は認証主体をUserとして冪等に登録するportです。
type UserRegistrar interface {
	Register(context.Context, RegisterUserParams) (*userdomain.User, bool, error)
}

// RegisteredUserResult はHTTPやDBに依存しないUser登録結果です。
type RegisteredUserResult struct {
	ID            string
	Email         string
	EmailVerified bool
	Role          string
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Created       bool
}

// RegisterAuthenticatedUserUseCase はAuth0のprofileをHogeDD Userへ紐づけます。
type RegisterAuthenticatedUserUseCase struct {
	profiles  AuthenticatedProfileProvider
	registrar UserRegistrar
}

// NewRegisterAuthenticatedUserUseCase はprofile取得とUser永続化のportを受け取ります。
func NewRegisterAuthenticatedUserUseCase(
	profiles AuthenticatedProfileProvider,
	registrar UserRegistrar,
) *RegisterAuthenticatedUserUseCase {
	return &RegisterAuthenticatedUserUseCase{profiles: profiles, registrar: registrar}
}

// Execute は認証主体をmemberとして冪等に登録し、最新のemail snapshotを返します。
// 既存Userのroleとstatusは変更しません。
func (uc *RegisterAuthenticatedUserUseCase) Execute(
	ctx context.Context,
	authenticated identity.Identity,
	accessToken string,
) (RegisteredUserResult, error) {
	if accessToken == "" {
		return RegisteredUserResult{}, ErrAccessTokenRequired
	}

	profile, err := uc.profiles.Fetch(ctx, accessToken)
	if err != nil {
		return RegisteredUserResult{}, fmt.Errorf("%w: %v", ErrProfileUnavailable, err)
	}
	if profile.Subject != authenticated.Subject() {
		return RegisteredUserResult{}, ErrProfileIdentityMismatch
	}

	registered, created, err := uc.registrar.Register(ctx, RegisterUserParams{
		AuthIssuer:    authenticated.Issuer(),
		AuthSubject:   authenticated.Subject(),
		Email:         profile.Email,
		EmailVerified: profile.EmailVerified,
		InitialRole:   userdomain.RoleMember,
	})
	if err != nil {
		return RegisteredUserResult{}, fmt.Errorf("%w: %v", ErrRegistrationFailed, err)
	}

	return toUserResult(registered, created), nil
}

func toUserResult(registered *userdomain.User, created bool) RegisteredUserResult {
	return RegisteredUserResult{
		ID:            registered.ID(),
		Email:         registered.Email(),
		EmailVerified: registered.EmailVerified(),
		Role:          registered.Role().String(),
		Status:        registered.Status().String(),
		CreatedAt:     registered.CreatedAt(),
		UpdatedAt:     registered.UpdatedAt(),
		Created:       created,
	}
}
