package domain

import "fmt"

// PublicationStatus はコンテンツを公開APIへ露出できる状態かを表します。
type PublicationStatus string

const (
	// PublicationStatusPreparing はコンテンツが公開準備中であることを表します。
	PublicationStatusPreparing PublicationStatus = "preparing"
	// PublicationStatusPublished はコンテンツが公開済みであることを表します。
	PublicationStatusPublished PublicationStatus = "published"
)

// IsPublic は公開APIから取得可能な状態の場合にtrueを返します。
func (s PublicationStatus) IsPublic() bool {
	return s == PublicationStatusPublished
}

// ParsePublicationStatus は永続化された文字列を検証済みPublicationStatusへ変換します。
func ParsePublicationStatus(value string) (PublicationStatus, error) {
	status := PublicationStatus(value)
	switch status {
	case PublicationStatusPreparing, PublicationStatusPublished:
		return status, nil
	default:
		return "", fmt.Errorf("invalid publication status: %q", value)
	}
}
