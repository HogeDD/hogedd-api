package domain

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	// MaxTitleLength はApp名の最大文字数です。
	MaxTitleLength = 100
	// MaxDescriptionLength はApp紹介文の最大文字数です。
	MaxDescriptionLength = 1000
	// MaxTags はAppへ設定できるタグ数です。
	MaxTags = 10
	// MaxTagLength は1タグの最大文字数です。
	MaxTagLength = 30
)

var (
	// ErrTitleRequired はAppの表示名が空であることを表します。
	ErrTitleRequired = errors.New("app title is required")
	// ErrDescriptionRequired はAppの紹介文が空であることを表します。
	ErrDescriptionRequired = errors.New("app description is required")
	// ErrTitleTooLong はApp名が上限を超えていることを表します。
	ErrTitleTooLong = errors.New("app title is too long")
	// ErrDescriptionTooLong はApp紹介文が上限を超えていることを表します。
	ErrDescriptionTooLong = errors.New("app description is too long")
	// ErrInvalidTags はタグが個数・文字数・一意性の制約を満たさないことを表します。
	ErrInvalidTags = errors.New("invalid app tags")
	// ErrPublishedAtRequired は公開日時が指定されていないことを表します。
	ErrPublishedAtRequired = errors.New("published at is required")
	// ErrDevelopmentDriveRequired は開発動機が空であることを表します。
	ErrDevelopmentDriveRequired = errors.New("development drive is required")
	// ErrInvalidYouTubeURL は紹介動画として利用できないURLであることを表します。
	ErrInvalidYouTubeURL = errors.New("invalid YouTube URL")
	// ErrPublishedAppCannotBeEdited は公開済みAppをDraft編集しようとしたことを表します。
	ErrPublishedAppCannotBeEdited = errors.New("published app cannot be edited")
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

// UpdateDraftDetails は公開準備中Appの表示情報を検証して更新します。
func (a *App) UpdateDraftDetails(title, description string, tags []string) error {
	if a.publicationStatus != PublicationStatusPreparing && a.publicationStatus != PublicationStatusPrivate {
		return ErrPublishedAppCannotBeEdited
	}
	updated, err := NewPreparingApp(a.slug, title, description, tags)
	if err != nil {
		return err
	}
	a.title = updated.title
	a.description = updated.description
	a.tags = updated.tags
	return nil
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
	if utf8.RuneCountInString(title) > MaxTitleLength {
		return nil, ErrTitleTooLong
	}

	description = strings.TrimSpace(description)
	if description == "" {
		return nil, ErrDescriptionRequired
	}
	if utf8.RuneCountInString(description) > MaxDescriptionLength {
		return nil, ErrDescriptionTooLong
	}
	if len(tags) > MaxTags {
		return nil, ErrInvalidTags
	}
	normalizedTags := make([]string, 0, len(tags))
	seenTags := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" || utf8.RuneCountInString(tag) > MaxTagLength {
			return nil, ErrInvalidTags
		}
		key := strings.ToLower(tag)
		if _, exists := seenTags[key]; exists {
			return nil, ErrInvalidTags
		}
		seenTags[key] = struct{}{}
		normalizedTags = append(normalizedTags, tag)
	}

	return &App{
		slug:              slug,
		title:             title,
		description:       description,
		tags:              normalizedTags,
		publicationStatus: PublicationStatusPreparing,
	}, nil
}

// RestoreApp は永続化された値からAppを復元し、domain invariantを再検証します。
func RestoreApp(slug Slug, title, description string, tags []string, status PublicationStatus, publishedAt time.Time, developmentDrive, youTubeURL string) (*App, error) {
	app, err := NewPreparingApp(slug, title, description, tags)
	if err != nil {
		return nil, err
	}
	switch status {
	case PublicationStatusPreparing:
		if !publishedAt.IsZero() || developmentDrive != "" || youTubeURL != "" {
			return nil, errors.New("preparing app has publication fields")
		}
	case PublicationStatusPublished:
		if err := app.Publish(publishedAt, developmentDrive, youTubeURL); err != nil {
			return nil, err
		}
	case PublicationStatusPrivate:
		if err := app.Publish(publishedAt, developmentDrive, youTubeURL); err != nil {
			return nil, err
		}
		app.publicationStatus = PublicationStatusPrivate
	default:
		return nil, errors.New("invalid publication status")
	}
	return app, nil
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

// MakePrivate は公開情報を保持したまま一般公開を停止します。
func (a *App) MakePrivate() error {
	if !a.publicationStatus.IsPublic() {
		return errors.New("only published app can be made private")
	}
	a.publicationStatus = PublicationStatusPrivate
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
