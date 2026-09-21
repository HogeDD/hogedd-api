package domain

import (
	"errors"
	"net/url"
	"strings"
	"time"
)

var (
	// ErrTitleRequired はAppの表示名が空であることを表します。
	ErrTitleRequired = errors.New("app title is required")
	// ErrDescriptionRequired はAppの紹介文が空であることを表します。
	ErrDescriptionRequired = errors.New("app description is required")
	// ErrPublishedAtRequired は公開日時が指定されていないことを表します。
	ErrPublishedAtRequired = errors.New("published at is required")
	// ErrDevelopmentDriveRequired は開発動機が空であることを表します。
	ErrDevelopmentDriveRequired = errors.New("development drive is required")
	// ErrInvalidYouTubeURL は紹介動画として利用できないURLであることを表します。
	ErrInvalidYouTubeURL = errors.New("invalid YouTube URL")
)

// App はHogeDDが公開するアプリと紹介動画の組み合わせを表します。
// Slugを同一性として持ち、公開に必要な情報が揃うまで公開済みにはなりません。
type App struct {
	slug              Slug
	title             string
	description       string
	tags              []string
	publicationStatus PublicationStatus
	publishedAt       time.Time
	developmentDrive  string
	youTubeURL        string
}

// NewPreparingApp は公開準備中のAppを生成します。
func NewPreparingApp(slug Slug, title, description string, tags []string) (*App, error) {
	if slug.String() == "" {
		return nil, ErrInvalidSlug
	}

	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrTitleRequired
	}

	description = strings.TrimSpace(description)
	if description == "" {
		return nil, ErrDescriptionRequired
	}

	return &App{
		slug:              slug,
		title:             title,
		description:       description,
		tags:              append([]string(nil), tags...),
		publicationStatus: PublicationStatusPreparing,
	}, nil
}

// Publish は公開に必要な情報を検証し、Appを公開済みにします。
func (a *App) Publish(publishedAt time.Time, developmentDrive, youTubeURL string) error {
	if publishedAt.IsZero() {
		return ErrPublishedAtRequired
	}

	developmentDrive = strings.TrimSpace(developmentDrive)
	if developmentDrive == "" {
		return ErrDevelopmentDriveRequired
	}

	normalizedYouTubeURL, err := normalizeYouTubeURL(youTubeURL)
	if err != nil {
		return err
	}

	a.publicationStatus = PublicationStatusPublished
	a.publishedAt = publishedAt
	a.developmentDrive = developmentDrive
	a.youTubeURL = normalizedYouTubeURL

	return nil
}

// Slug はAppを識別するSlugを返します。
func (a *App) Slug() Slug {
	return a.slug
}

// Title は表示用のアプリ名を返します。
func (a *App) Title() string {
	return a.title
}

// Description はアプリの紹介文を返します。
func (a *App) Description() string {
	return a.description
}

// Tags は呼び出し元から変更できないよう複製したタグ一覧を返します。
func (a *App) Tags() []string {
	return append([]string(nil), a.tags...)
}

// PublicationStatus はAppの公開状態を返します。
func (a *App) PublicationStatus() PublicationStatus {
	return a.publicationStatus
}

// PublishedAt は公開日時を返します。準備中の場合はゼロ値です。
func (a *App) PublishedAt() time.Time {
	return a.publishedAt
}

// DevelopmentDrive は開発動機を返します。準備中の場合は空文字です。
func (a *App) DevelopmentDrive() string {
	return a.developmentDrive
}

// YouTubeURL は正規化済みの紹介動画URLを返します。準備中の場合は空文字です。
func (a *App) YouTubeURL() string {
	return a.youTubeURL
}

func normalizeYouTubeURL(value string) (string, error) {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err != nil || parsed.Scheme != "https" {
		return "", ErrInvalidYouTubeURL
	}

	host := strings.ToLower(parsed.Hostname())
	if host != "youtube.com" && host != "www.youtube.com" && host != "youtu.be" {
		return "", ErrInvalidYouTubeURL
	}

	return parsed.String(), nil
}
