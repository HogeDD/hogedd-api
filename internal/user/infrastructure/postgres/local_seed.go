package postgres

import (
	"context"
	"database/sql"
	"fmt"

	userdomain "github.com/iwasawa/hogedd-api/internal/user/domain"
)

const localSeedIssuer = "https://seed.hogedd.invalid/"

// LocalSeedIdentity はLocal Auth0ユーザーとDB Userを紐づけるseed入力です。
type LocalSeedIdentity struct {
	// AuthIssuer はAccess Tokenのissuerと一致する値です。
	AuthIssuer string
	// AuthSubject はAccess Tokenのsubjectと一致する値です。
	AuthSubject string
	// Email はLocal環境で連絡先として保存するemailです。
	Email string
	// DisplayName はLocal環境で表示するプロフィール名です。
	DisplayName string
}

type localSeedUser struct {
	id            string
	authIssuer    string
	authSubject   string
	email         string
	emailVerified bool
	role          userdomain.Role
	status        userdomain.Status
	displayName   string
}

type seedExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func defaultLocalSeedUsers() []localSeedUser {
	return []localSeedUser{
		{
			id:            "00000000-0000-4000-8000-000000000001",
			authIssuer:    localSeedIssuer,
			authSubject:   "seed|owner-active",
			email:         "owner@seed.hogedd.invalid",
			emailVerified: true,
			role:          userdomain.RoleOwner,
			status:        userdomain.StatusActive,
			displayName:   "Seed Owner",
		},
		{
			id:            "00000000-0000-4000-8000-000000000002",
			authIssuer:    localSeedIssuer,
			authSubject:   "seed|admin-active",
			email:         "admin@seed.hogedd.invalid",
			emailVerified: true,
			role:          userdomain.RoleAdmin,
			status:        userdomain.StatusActive,
			displayName:   "Seed Admin",
		},
		{
			id:            "00000000-0000-4000-8000-000000000003",
			authIssuer:    localSeedIssuer,
			authSubject:   "seed|member-active",
			email:         "member@seed.hogedd.invalid",
			emailVerified: true,
			role:          userdomain.RoleMember,
			status:        userdomain.StatusActive,
			displayName:   "Seed Member",
		},
		{
			id:            "00000000-0000-4000-8000-000000000004",
			authIssuer:    localSeedIssuer,
			authSubject:   "seed|member-disabled",
			email:         "disabled@seed.hogedd.invalid",
			emailVerified: false,
			role:          userdomain.RoleMember,
			status:        userdomain.StatusDisabled,
		},
	}
}

// SeedLocalDevelopment はLocal DBへ代表的なUser状態を冪等に保存します。
func SeedLocalDevelopment(ctx context.Context, db *sql.DB, identity *LocalSeedIdentity) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin local seed transaction: %w", err)
	}
	defer tx.Rollback()

	users := defaultLocalSeedUsers()
	if identity != nil {
		users = append(users, localSeedUser{
			authIssuer:    identity.AuthIssuer,
			authSubject:   identity.AuthSubject,
			email:         identity.Email,
			emailVerified: true,
			role:          userdomain.RoleOwner,
			status:        userdomain.StatusActive,
			displayName:   identity.DisplayName,
		})
	}
	if err := seedLocalUsers(ctx, tx, users); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit local seed transaction: %w", err)
	}
	return nil
}

func seedLocalUsers(ctx context.Context, executor seedExecutor, users []localSeedUser) error {
	for _, user := range users {
		var err error
		if user.id == "" {
			_, err = executor.ExecContext(ctx, `
INSERT INTO users (auth_issuer, auth_subject, email, email_verified, role, status)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (auth_issuer, auth_subject) DO UPDATE SET
    email = EXCLUDED.email,
    email_verified = EXCLUDED.email_verified,
    role = EXCLUDED.role,
    status = EXCLUDED.status,
    updated_at = CURRENT_TIMESTAMP`,
				user.authIssuer,
				user.authSubject,
				user.email,
				user.emailVerified,
				user.role.String(),
				user.status.String(),
			)
		} else {
			_, err = executor.ExecContext(ctx, `
INSERT INTO users (id, auth_issuer, auth_subject, email, email_verified, role, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (auth_issuer, auth_subject) DO UPDATE SET
    email = EXCLUDED.email,
    email_verified = EXCLUDED.email_verified,
    role = EXCLUDED.role,
    status = EXCLUDED.status,
    updated_at = CURRENT_TIMESTAMP`,
				user.id,
				user.authIssuer,
				user.authSubject,
				user.email,
				user.emailVerified,
				user.role.String(),
				user.status.String(),
			)
		}
		if err != nil {
			return fmt.Errorf("upsert local seed user %q: %w", user.authSubject, err)
		}

		if user.displayName == "" {
			if _, err := executor.ExecContext(ctx, `
DELETE FROM user_profiles
WHERE user_id = (SELECT id FROM users WHERE auth_issuer = $1 AND auth_subject = $2)`, user.authIssuer, user.authSubject); err != nil {
				return fmt.Errorf("delete local seed profile %q: %w", user.authSubject, err)
			}
			continue
		}
		if _, err := executor.ExecContext(ctx, `
INSERT INTO user_profiles (user_id, display_name)
SELECT id, $3 FROM users WHERE auth_issuer = $1 AND auth_subject = $2
ON CONFLICT (user_id) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    updated_at = CURRENT_TIMESTAMP`, user.authIssuer, user.authSubject, user.displayName); err != nil {
			return fmt.Errorf("upsert local seed profile %q: %w", user.authSubject, err)
		}
	}
	return nil
}
