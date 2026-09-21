package domain

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
