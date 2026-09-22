package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	userapp "github.com/iwasawa/hogedd-api/internal/user/application"
	userdomain "github.com/iwasawa/hogedd-api/internal/user/domain"
)

type rowScanner interface {
	Scan(...any) error
}

type queryRowFunc func(context.Context, string, ...any) rowScanner

// UserRegistrar は認証主体をPostgreSQLのusersへ冪等に保存します。
type UserRegistrar struct {
	queryRow queryRowFunc
}

// NewUserRegistrar はDB接続を使うUserRegistrarを構築します。
func NewUserRegistrar(db *sql.DB) *UserRegistrar {
	return &UserRegistrar{queryRow: func(ctx context.Context, query string, args ...any) rowScanner {
		return db.QueryRowContext(ctx, query, args...)
	}}
}

// Register は未登録ならmemberを作成し、登録済みならemail snapshotだけを更新します。
// 既存Userのroleとstatusは変更しません。
func (r *UserRegistrar) Register(
	ctx context.Context,
	params userapp.RegisterUserParams,
) (*userdomain.User, bool, error) {
	const insertQuery = `
INSERT INTO users (auth_issuer, auth_subject, email, email_verified, role)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (auth_issuer, auth_subject) DO NOTHING
RETURNING id, auth_issuer, auth_subject, email, email_verified, role, status, created_at, updated_at`

	registered, err := scanUser(r.queryRow(
		ctx,
		insertQuery,
		params.AuthIssuer,
		params.AuthSubject,
		params.Email,
		params.EmailVerified,
		params.InitialRole.String(),
	))
	if err == nil {
		return registered, true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, false, fmt.Errorf("insert user: %w", err)
	}

	const updateQuery = `
UPDATE users
SET email = $3, email_verified = $4, updated_at = CURRENT_TIMESTAMP
WHERE auth_issuer = $1 AND auth_subject = $2
RETURNING id, auth_issuer, auth_subject, email, email_verified, role, status, created_at, updated_at`
	registered, err = scanUser(r.queryRow(
		ctx,
		updateQuery,
		params.AuthIssuer,
		params.AuthSubject,
		params.Email,
		params.EmailVerified,
	))
	if err != nil {
		return nil, false, fmt.Errorf("update user contact snapshot: %w", err)
	}
	return registered, false, nil
}

func scanUser(row rowScanner) (*userdomain.User, error) {
	var (
		id            string
		authIssuer    string
		authSubject   string
		email         string
		emailVerified bool
		roleValue     string
		statusValue   string
		createdAt     sql.NullTime
		updatedAt     sql.NullTime
	)
	if err := row.Scan(
		&id,
		&authIssuer,
		&authSubject,
		&email,
		&emailVerified,
		&roleValue,
		&statusValue,
		&createdAt,
		&updatedAt,
	); err != nil {
		return nil, err
	}
	role, err := userdomain.ParseRole(roleValue)
	if err != nil {
		return nil, err
	}
	status, err := userdomain.ParseStatus(statusValue)
	if err != nil {
		return nil, err
	}
	return userdomain.RestoreUser(
		id,
		authIssuer,
		authSubject,
		email,
		emailVerified,
		role,
		status,
		createdAt.Time,
		updatedAt.Time,
	)
}
