package domain

import (
	"errors"
	"regexp"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// ErrInvalidSlug はSlugに利用できない文字列が指定されたことを表します。
var ErrInvalidSlug = errors.New("invalid content slug")

// Slug は公開コンテンツをURL上で一意に識別する値です。
// 小文字の英数字をハイフンで区切った形式だけを許可します。
type Slug struct {
	value string
}

// NewSlug は検証済みのSlugを生成します。
func NewSlug(value string) (Slug, error) {
	if !slugPattern.MatchString(value) {
		return Slug{}, ErrInvalidSlug
	}

	return Slug{value: value}, nil
}

// String はURLやレスポンスへ使用できるSlugの文字列表現を返します。
func (s Slug) String() string {
	return s.value
}
