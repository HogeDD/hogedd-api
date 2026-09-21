package infrastructure

import (
	"fmt"
	"time"

	"github.com/iwasawa/hogedd-api/internal/content/domain"
)

type appSeed struct {
	slug             string
	title            string
	description      string
	tags             []string
	publishedAt      string
	developmentDrive string
	youTubeURL       string
}

var initialApps = []appSeed{
	{
		slug:             "clean-tasks",
		title:            "Clean Tasks",
		description:      "Clean Architecture の練習として作った、最小構成のタスクアプリ。",
		tags:             []string{"Next.js", "TypeScript", "Clean Architecture"},
		publishedAt:      "2026-05-30",
		developmentDrive: "学習DD",
		youTubeURL:       "https://youtu.be/5mo0qnPuVTY?si=5comgyazMjAYnOJ9",
	},
	{
		slug:        "chinchin",
		title:       "ちんちんゲーム",
		description: "5x5 の盤面で「ち」と「ん」を交互に置く2人対戦ゲーム。",
		tags:        []string{"Game", "Local Multiplayer", "Next.js"},
	},
	{
		slug:        "nishida",
		title:       "ニシ打",
		description: "ラランド・ニシダにまつわる語彙で遊ぶ、ローマ字タイピングゲーム。",
		tags:        []string{"Game", "Typing", "Next.js"},
	},
	{
		slug:        "judo-roulette",
		title:       "柔道ルーレット",
		description: "選択肢を追加して回せる、仕込み可能なルーレット。",
		tags:        []string{"Game", "Roulette", "Next.js"},
	},
	{
		slug:        "bakuon-kikiippatsu",
		title:       "爆音危機一髪",
		description: "1つだけ爆音が鳴るボタンを避けながら、みんなで順番に押していくゲーム。",
		tags:        []string{"Game", "Local Multiplayer", "Sound"},
	},
}

// NewSeededMemoryAppStore はhogedd-webの現行アプリ定義から取得元を構築します。
// publishedAtが空のAppは公開準備中として扱います。
func NewSeededMemoryAppStore() (*MemoryAppStore, error) {
	apps := make([]*domain.App, 0, len(initialApps))
	for _, seed := range initialApps {
		slug, err := domain.NewSlug(seed.slug)
		if err != nil {
			return nil, fmt.Errorf("seed app %q: %w", seed.slug, err)
		}
		app, err := domain.NewPreparingApp(slug, seed.title, seed.description, seed.tags)
		if err != nil {
			return nil, fmt.Errorf("seed app %q: %w", seed.slug, err)
		}
		if seed.publishedAt != "" {
			publishedAt, err := time.Parse("2006-01-02", seed.publishedAt)
			if err != nil {
				return nil, fmt.Errorf("seed app %q: parse publication date: %w", seed.slug, err)
			}
			if err := app.Publish(publishedAt, seed.developmentDrive, seed.youTubeURL); err != nil {
				return nil, fmt.Errorf("seed app %q: %w", seed.slug, err)
			}
		}
		apps = append(apps, app)
	}
	return NewMemoryAppStore(apps), nil
}
