package domain

import (
	"crypto/rand"
	"fmt"
)

// AppID はContent Appの変更されない同一性を表すUUIDです。
type AppID string

// NewAppID は暗号学的乱数からUUID v4形式のAppIDを生成します。
func NewAppID() (AppID, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("generate app id: %w", err)
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return AppID(fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", value[0:4], value[4:6], value[6:8], value[8:10], value[10:16])), nil
}

// RestoreAppID は永続化済みUUIDをAppIDへ復元します。
func RestoreAppID(value string) (AppID, error) {
	if len(value) != 36 {
		return "", fmt.Errorf("invalid app id: %q", value)
	}
	return AppID(value), nil
}

// String はUUID文字列を返します。
func (id AppID) String() string { return string(id) }
