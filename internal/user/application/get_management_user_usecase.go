package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/iwasawa/hogedd-api/internal/identity"
	userdomain "github.com/iwasawa/hogedd-api/internal/user/domain"
)

var (
	// ErrManagementUserNotFound は認証主体に紐づく運営Userが存在しないことを表します。
	ErrManagementUserNotFound = errors.New("management user not found")
	// ErrManagementUserInactive は運営Userが利用停止中であることを表します。
	ErrManagementUserInactive = errors.New("management user is inactive")
	// ErrManagementPermissionDenied はUserが運営権限を持たないことを表します。
	ErrManagementPermissionDenied = errors.New("management permission denied")
)

// ManagementUserResult は運営境界を通過したUserの最小情報です。
type ManagementUserResult struct {
	ID   string
	Role string
}

// GetManagementUserUseCase は認証主体が運営機能を利用できるか判定します。
type GetManagementUserUseCase struct {
	finder UserFinder
}

// NewGetManagementUserUseCase はUser取得portを使う運営認可Use Caseを構築します。
func NewGetManagementUserUseCase(finder UserFinder) *GetManagementUserUseCase {
	return &GetManagementUserUseCase{finder: finder}
}

// Execute はactiveなowner・adminだけを運営Userとして返します。
func (uc *GetManagementUserUseCase) Execute(ctx context.Context, authenticated identity.Identity) (ManagementUserResult, error) {
	user, found, err := uc.finder.FindByIdentity(ctx, authenticated.Issuer(), authenticated.Subject())
	if err != nil {
		return ManagementUserResult{}, fmt.Errorf("find management user: %w", err)
	}
	if !found {
		return ManagementUserResult{}, ErrManagementUserNotFound
	}
	if user.Status() != userdomain.StatusActive {
		return ManagementUserResult{}, ErrManagementUserInactive
	}
	if !user.Role().AllowsManagement() {
		return ManagementUserResult{}, ErrManagementPermissionDenied
	}
	return ManagementUserResult{ID: user.ID(), Role: user.Role().String()}, nil
}
