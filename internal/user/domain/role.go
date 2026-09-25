package domain

import "fmt"

// Role はHogeDD内でUserに許可される権限区分です。
type Role string

const (
	// RoleOwner はHogeDD全体を管理する最上位の権限です。
	RoleOwner Role = "owner"
	// RoleAdmin は日常の運用管理を行う権限です。
	RoleAdmin Role = "admin"
	// RoleMember は自己登録時に付与する通常利用者の権限です。
	RoleMember Role = "member"
)

// ParseRole は永続化された文字列を検証済みRoleへ変換します。
func ParseRole(value string) (Role, error) {
	role := Role(value)
	switch role {
	case RoleOwner, RoleAdmin, RoleMember:
		return role, nil
	default:
		return "", fmt.Errorf("invalid user role: %q", value)
	}
}

// String はRoleの永続化・表示用文字列を返します。
func (r Role) String() string {
	return string(r)
}

// AllowsManagement は運営機能を利用できるRoleかを返します。
func (r Role) AllowsManagement() bool {
	return r == RoleOwner || r == RoleAdmin
}
