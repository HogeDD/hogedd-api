package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewPreparingApp(t *testing.T) {
	t.Parallel()

	slug := mustSlug(t, "clean-tasks")
	tags := []string{"Next.js", "TypeScript"}

	app, err := NewPreparingApp(slug, " Clean Tasks ", " タスク管理アプリ。 ", tags)
	if err != nil {
		t.Fatalf("NewPreparingApp() unexpected error = %v", err)
	}

	if app.Slug() != slug {
		t.Fatalf("Slug() = %v, want %v", app.Slug(), slug)
	}
	if app.Title() != "Clean Tasks" {
		t.Fatalf("Title() = %q, want %q", app.Title(), "Clean Tasks")
	}
	if app.Description() != "タスク管理アプリ。" {
		t.Fatalf("Description() = %q, want %q", app.Description(), "タスク管理アプリ。")
	}
	if app.PublicationStatus() != PublicationStatusPreparing {
		t.Fatalf("PublicationStatus() = %q, want %q", app.PublicationStatus(), PublicationStatusPreparing)
	}

	tags[0] = "changed"
	gotTags := app.Tags()
	gotTags[1] = "changed"
	if app.Tags()[0] != "Next.js" || app.Tags()[1] != "TypeScript" {
		t.Fatal("Tags() allowed mutation from outside the entity")
	}
}

func TestNewPreparingAppRejectsMissingText(t *testing.T) {
	t.Parallel()

	slug := mustSlug(t, "clean-tasks")
	tests := []struct {
		name        string
		title       string
		description string
		wantErr     error
	}{
		{name: "タイトルなし", description: "説明", wantErr: ErrTitleRequired},
		{name: "空白だけのタイトル", title: "  ", description: "説明", wantErr: ErrTitleRequired},
		{name: "説明なし", title: "名前", wantErr: ErrDescriptionRequired},
		{name: "空白だけの説明", title: "名前", description: "  ", wantErr: ErrDescriptionRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewPreparingApp(slug, tt.title, tt.description, nil)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewPreparingApp() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewPreparingAppRejectsEmptySlug(t *testing.T) {
	t.Parallel()

	_, err := NewPreparingApp(Slug{}, "名前", "説明", nil)
	if !errors.Is(err, ErrInvalidSlug) {
		t.Fatalf("NewPreparingApp() error = %v, want %v", err, ErrInvalidSlug)
	}
}

func TestAppPublish(t *testing.T) {
	t.Parallel()

	app := mustPreparingApp(t)
	publishedAt := time.Date(2026, time.May, 30, 0, 0, 0, 0, time.UTC)

	err := app.Publish(
		publishedAt,
		" 学習DD ",
		"https://youtu.be/5mo0qnPuVTY?si=example",
	)
	if err != nil {
		t.Fatalf("Publish() unexpected error = %v", err)
	}

	if !app.PublicationStatus().IsPublic() {
		t.Fatalf("PublicationStatus() = %q, want public", app.PublicationStatus())
	}
	if !app.PublishedAt().Equal(publishedAt) {
		t.Fatalf("PublishedAt() = %v, want %v", app.PublishedAt(), publishedAt)
	}
	if app.DevelopmentDrive() != "学習DD" {
		t.Fatalf("DevelopmentDrive() = %q, want %q", app.DevelopmentDrive(), "学習DD")
	}
	if app.YouTubeURL() != "https://youtu.be/5mo0qnPuVTY?si=example" {
		t.Fatalf("YouTubeURL() = %q", app.YouTubeURL())
	}
}

func TestAppPublishRejectsIncompletePublication(t *testing.T) {
	t.Parallel()

	publishedAt := time.Date(2026, time.May, 30, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name             string
		publishedAt      time.Time
		developmentDrive string
		youTubeURL       string
		wantErr          error
	}{
		{name: "公開日時なし", developmentDrive: "学習DD", youTubeURL: "https://youtu.be/video", wantErr: ErrPublishedAtRequired},
		{name: "開発動機なし", publishedAt: publishedAt, youTubeURL: "https://youtu.be/video", wantErr: ErrDevelopmentDriveRequired},
		{name: "URLではない", publishedAt: publishedAt, developmentDrive: "学習DD", youTubeURL: "video", wantErr: ErrInvalidYouTubeURL},
		{name: "HTTP", publishedAt: publishedAt, developmentDrive: "学習DD", youTubeURL: "http://youtu.be/video", wantErr: ErrInvalidYouTubeURL},
		{name: "YouTube以外", publishedAt: publishedAt, developmentDrive: "学習DD", youTubeURL: "https://example.com/video", wantErr: ErrInvalidYouTubeURL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := mustPreparingApp(t)
			err := app.Publish(tt.publishedAt, tt.developmentDrive, tt.youTubeURL)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Publish() error = %v, want %v", err, tt.wantErr)
			}
			if app.PublicationStatus() != PublicationStatusPreparing {
				t.Fatalf("failed Publish() changed status to %q", app.PublicationStatus())
			}
		})
	}
}

func mustSlug(t *testing.T, value string) Slug {
	t.Helper()

	slug, err := NewSlug(value)
	if err != nil {
		t.Fatalf("NewSlug() unexpected error = %v", err)
	}
	return slug
}

func mustPreparingApp(t *testing.T) *App {
	t.Helper()

	app, err := NewPreparingApp(
		mustSlug(t, "clean-tasks"),
		"Clean Tasks",
		"タスク管理アプリ。",
		[]string{"Next.js"},
	)
	if err != nil {
		t.Fatalf("NewPreparingApp() unexpected error = %v", err)
	}
	return app
}
