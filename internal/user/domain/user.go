package domain

import (
	"errors"
	"net/mail"
	"strings"
	"time"
)

var (
	// ErrUserIDRequired はHogeDD内部のUser IDが空であることを表します。
	ErrUserIDRequired = errors.New("user ID is required")
	// ErrAuthIssuerRequired は認証主体のissuerが空であることを表します。
	ErrAuthIssuerRequired = errors.New("auth issuer is required")
	// ErrAuthSubjectRequired は認証主体のsubjectが空であることを表します。
	ErrAuthSubjectRequired = errors.New("auth subject is required")
	// ErrInvalidEmail は通知先として利用できないemailであることを表します。
	ErrInvalidEmail = errors.New("invalid user email")
)

// User はAuth0の認証主体とHogeDD内の権限を結びつける利用者です。
// 同一性はauthIssuerとauthSubjectの組み合わせで決まり、email変更には影響されません。
type User struct {
	id            string
	authIssuer    string
	authSubject   string
	email         string
	emailVerified bool
	role          Role
	status        Status
	createdAt     time.Time
	updatedAt     time.Time
}

// RestoreUser は永続化済みの値を検証し、Userを復元します。
func RestoreUser(
	id string,
	authIssuer string,
	authSubject string,
	email string,
	emailVerified bool,
	role Role,
	status Status,
	createdAt time.Time,
	updatedAt time.Time,
) (*User, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrUserIDRequired
	}
	if strings.TrimSpace(authIssuer) == "" {
		return nil, ErrAuthIssuerRequired
	}
	if strings.TrimSpace(authSubject) == "" {
		return nil, ErrAuthSubjectRequired
	}

	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return nil, err
	}
	if _, err := ParseRole(role.String()); err != nil {
		return nil, err
	}
	if _, err := ParseStatus(status.String()); err != nil {
		return nil, err
	}

	return &User{
		id:            id,
		authIssuer:    authIssuer,
		authSubject:   authSubject,
		email:         normalizedEmail,
		emailVerified: emailVerified,
		role:          role,
		status:        status,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}, nil
}

func normalizeEmail(value string) (string, error) {
	value = strings.TrimSpace(value)
	parsed, err := mail.ParseAddress(value)
	if err != nil || !strings.EqualFold(parsed.Address, value) {
		return "", ErrInvalidEmail
	}
	return strings.ToLower(parsed.Address), nil
}

// ID はHogeDD内部でUserを識別するopaqueなIDを返します。
func (u *User) ID() string { return u.id }

// AuthIssuer は認証主体を発行したissuerを返します。
func (u *User) AuthIssuer() string { return u.authIssuer }

// AuthSubject はissuer内で認証主体を識別するsubjectを返します。
func (u *User) AuthSubject() string { return u.authSubject }

// Email は通知先として保存したemail snapshotを返します。
func (u *User) Email() string { return u.email }

// EmailVerified はemailがAuth0で確認済みかを返します。
func (u *User) EmailVerified() bool { return u.emailVerified }

// Role はHogeDD内での権限を返します。
func (u *User) Role() Role { return u.role }

// Status はHogeDDを利用できる状態かを返します。
func (u *User) Status() Status { return u.status }

// CreatedAt はUserが最初に登録された時刻を返します。
func (u *User) CreatedAt() time.Time { return u.createdAt }

// UpdatedAt はUserのsnapshotが最後に更新された時刻を返します。
func (u *User) UpdatedAt() time.Time { return u.updatedAt }
