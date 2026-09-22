package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	userapp "github.com/iwasawa/hogedd-api/internal/user/application"
	userdomain "github.com/iwasawa/hogedd-api/internal/user/domain"
)

// ProfileRepository はUserプロフィールをPostgreSQLへ保存します。
type ProfileRepository struct {
	queryRow func(context.Context, string, ...any) rowScanner
}

// NewProfileRepository はProfileRepositoryを構築します。
func NewProfileRepository(db *sql.DB) *ProfileRepository {
	return &ProfileRepository{queryRow: func(ctx context.Context, query string, args ...any) rowScanner {
		return db.QueryRowContext(ctx, query, args...)
	}}
}

// FindByIdentity は認証主体に紐づくUser状態とプロフィールを返します。
func (r *ProfileRepository) FindByIdentity(ctx context.Context, issuer, subject string) (*userdomain.Profile, userdomain.Status, bool, error) {
	const query = `SELECT u.id, u.status, p.display_name
FROM users u LEFT JOIN user_profiles p ON p.user_id = u.id
WHERE u.auth_issuer = $1 AND u.auth_subject = $2`
	var userID, statusValue string
	var displayName sql.NullString
	if err := r.queryRow(ctx, query, issuer, subject).Scan(&userID, &statusValue, &displayName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", false, nil
		}
		return nil, "", false, fmt.Errorf("query user profile: %w", err)
	}
	status, err := userdomain.ParseStatus(statusValue)
	if err != nil {
		return nil, "", false, err
	}
	if !displayName.Valid {
		return nil, status, false, nil
	}
	profile, err := userdomain.NewProfile(userID, displayName.String)
	return profile, status, true, err
}

// SaveByIdentity は認証主体に紐づくUserの表示名をupsertします。
func (r *ProfileRepository) SaveByIdentity(ctx context.Context, issuer, subject, displayName string) (*userdomain.Profile, userdomain.Status, error) {
	const query = `WITH matched_user AS (
  SELECT id, status FROM users WHERE auth_issuer = $1 AND auth_subject = $2
), saved AS (
  INSERT INTO user_profiles (user_id, display_name)
  SELECT id, $3 FROM matched_user WHERE status = 'active'
  ON CONFLICT (user_id) DO UPDATE SET display_name = EXCLUDED.display_name, updated_at = CURRENT_TIMESTAMP
  RETURNING user_id, display_name
)
SELECT matched_user.id, matched_user.status, saved.display_name
FROM matched_user LEFT JOIN saved ON saved.user_id = matched_user.id`
	var userID, statusValue string
	var savedName sql.NullString
	if err := r.queryRow(ctx, query, issuer, subject, displayName).Scan(&userID, &statusValue, &savedName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", userapp.ErrUserNotFound
		}
		return nil, "", fmt.Errorf("upsert user profile: %w", err)
	}
	status, err := userdomain.ParseStatus(statusValue)
	if err != nil {
		return nil, "", err
	}
	if !savedName.Valid {
		return nil, status, nil
	}
	profile, err := userdomain.NewProfile(userID, savedName.String)
	return profile, status, err
}
