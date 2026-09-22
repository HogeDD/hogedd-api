package domain

import (
	"errors"
	"strings"
	"unicode/utf8"
)

const MaxDisplayNameLength = 50

var (
	// ErrDisplayNameRequired は表示名が空であることを表します。
	ErrDisplayNameRequired = errors.New("display name is required")
	// ErrDisplayNameTooLong は表示名が許容文字数を超えていることを表します。
	ErrDisplayNameTooLong = errors.New("display name is too long")
)

// Profile はHogeDD内で利用者が編集できるプロフィールです。
type Profile struct {
	userID      string
	displayName string
}

// NewProfile はUser IDと表示名を検証してProfileを構築します。
func NewProfile(userID, displayName string) (*Profile, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUserIDRequired
	}
	normalized := strings.Join(strings.Fields(displayName), " ")
	if normalized == "" {
		return nil, ErrDisplayNameRequired
	}
	if utf8.RuneCountInString(normalized) > MaxDisplayNameLength {
		return nil, ErrDisplayNameTooLong
	}
	return &Profile{userID: userID, displayName: normalized}, nil
}

// UserID はプロフィールを所有するHogeDD User IDを返します。
func (p *Profile) UserID() string { return p.userID }

// DisplayName は正規化済みの表示名を返します。
func (p *Profile) DisplayName() string { return p.displayName }
