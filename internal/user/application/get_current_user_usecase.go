package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/iwasawa/hogedd-api/internal/identity"
	userdomain "github.com/iwasawa/hogedd-api/internal/user/domain"
)

// ErrUserNotFound は認証主体に紐づくHogeDD Userが未登録であることを表します。
var ErrUserNotFound = errors.New("user not found")

// UserFinder は認証主体によるUser取得に必要なportです。
type UserFinder interface {
	FindByIdentity(context.Context, string, string) (*userdomain.User, bool, error)
}

// GetCurrentUserUseCase は認証主体に紐づくHogeDD Userを取得します。
type GetCurrentUserUseCase struct {
	finder UserFinder
}

// NewGetCurrentUserUseCase はUser取得portを使うUse Caseを構築します。
func NewGetCurrentUserUseCase(finder UserFinder) *GetCurrentUserUseCase {
	return &GetCurrentUserUseCase{finder: finder}
}

// Execute は検証済みIdentityに対応するUserを返します。
func (uc *GetCurrentUserUseCase) Execute(
	ctx context.Context,
	authenticated identity.Identity,
) (RegisteredUserResult, error) {
	user, found, err := uc.finder.FindByIdentity(ctx, authenticated.Issuer(), authenticated.Subject())
	if err != nil {
		return RegisteredUserResult{}, fmt.Errorf("find current user: %w", err)
	}
	if !found {
		return RegisteredUserResult{}, ErrUserNotFound
	}
	return toUserResult(user, false), nil
}
