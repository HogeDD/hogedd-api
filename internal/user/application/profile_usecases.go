package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/iwasawa/hogedd-api/internal/identity"
	userdomain "github.com/iwasawa/hogedd-api/internal/user/domain"
)

var (
	// ErrProfileNotFound は現在Userのプロフィールが未登録であることを表します。
	ErrProfileNotFound = errors.New("user profile not found")
	// ErrUserDisabled は現在Userが無効化されていることを表します。
	ErrUserDisabled = errors.New("user is disabled")
)

// UserProfileRepository は現在Userのプロフィール取得・保存に必要なportです。
type UserProfileRepository interface {
	FindByIdentity(context.Context, string, string) (*userdomain.Profile, userdomain.Status, bool, error)
	SaveByIdentity(context.Context, string, string, string) (*userdomain.Profile, userdomain.Status, error)
}

// ProfileResult はHTTPやDBに依存しないプロフィール結果です。
type ProfileResult struct {
	DisplayName string
}

// GetCurrentProfileUseCase は現在Userのプロフィールを取得します。
type GetCurrentProfileUseCase struct{ profiles UserProfileRepository }

// NewGetCurrentProfileUseCase はプロフィール取得Use Caseを構築します。
func NewGetCurrentProfileUseCase(profiles UserProfileRepository) *GetCurrentProfileUseCase {
	return &GetCurrentProfileUseCase{profiles: profiles}
}

// Execute はactiveな現在Userの登録済みプロフィールを返します。
func (uc *GetCurrentProfileUseCase) Execute(ctx context.Context, authenticated identity.Identity) (ProfileResult, error) {
	profile, status, found, err := uc.profiles.FindByIdentity(ctx, authenticated.Issuer(), authenticated.Subject())
	if err != nil {
		return ProfileResult{}, fmt.Errorf("find current profile: %w", err)
	}
	if status == "" {
		return ProfileResult{}, ErrUserNotFound
	}
	if status != userdomain.StatusActive {
		return ProfileResult{}, ErrUserDisabled
	}
	if !found {
		return ProfileResult{}, ErrProfileNotFound
	}
	return ProfileResult{DisplayName: profile.DisplayName()}, nil
}

// UpdateCurrentProfileUseCase は現在Userの表示名を登録・更新します。
type UpdateCurrentProfileUseCase struct{ profiles UserProfileRepository }

// NewUpdateCurrentProfileUseCase はプロフィール更新Use Caseを構築します。
func NewUpdateCurrentProfileUseCase(profiles UserProfileRepository) *UpdateCurrentProfileUseCase {
	return &UpdateCurrentProfileUseCase{profiles: profiles}
}

// Execute はactiveな現在Userの表示名を検証して保存します。
func (uc *UpdateCurrentProfileUseCase) Execute(ctx context.Context, authenticated identity.Identity, displayName string) (ProfileResult, error) {
	validated, err := userdomain.NewProfile("pending", displayName)
	if err != nil {
		return ProfileResult{}, err
	}
	profile, status, err := uc.profiles.SaveByIdentity(ctx, authenticated.Issuer(), authenticated.Subject(), validated.DisplayName())
	if errors.Is(err, ErrUserNotFound) {
		return ProfileResult{}, ErrUserNotFound
	}
	if err != nil {
		return ProfileResult{}, fmt.Errorf("save current profile: %w", err)
	}
	if status != userdomain.StatusActive {
		return ProfileResult{}, ErrUserDisabled
	}
	return ProfileResult{DisplayName: profile.DisplayName()}, nil
}
