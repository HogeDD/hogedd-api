package domain

import "fmt"

// Status はUserがHogeDDを利用できる状態かを表します。
type Status string

const (
	// StatusActive は利用可能なUserを表します。
	StatusActive Status = "active"
	// StatusDisabled は利用を停止されたUserを表します。
	StatusDisabled Status = "disabled"
)

// ParseStatus は永続化された文字列を検証済みStatusへ変換します。
func ParseStatus(value string) (Status, error) {
	status := Status(value)
	switch status {
	case StatusActive, StatusDisabled:
		return status, nil
	default:
		return "", fmt.Errorf("invalid user status: %q", value)
	}
}

// String はStatusの永続化・表示用文字列を返します。
func (s Status) String() string {
	return string(s)
}
