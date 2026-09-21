package infrastructure

import (
	"context"

	"github.com/iwasawa/hogedd-api/internal/content/domain"
)

// MemoryAppStore はAppをプロセス内のメモリから取得します。
// 公開判定は行わず、呼び出し元が内部状態を変更できないようAppを複製して返します。
type MemoryAppStore struct {
	apps []domain.App
}

// NewMemoryAppStore は渡されたAppを保持する取得元を構築します。
func NewMemoryAppStore(apps []*domain.App) *MemoryAppStore {
	stored := make([]domain.App, len(apps))
	for i, app := range apps {
		stored[i] = *app
	}
	return &MemoryAppStore{apps: stored}
}

// List は保存されているすべてのAppを順序を保って返します。
func (s *MemoryAppStore) List(ctx context.Context) ([]*domain.App, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	apps := make([]*domain.App, len(s.apps))
	for i := range s.apps {
		app := s.apps[i]
		apps[i] = &app
	}
	return apps, nil
}

// FindBySlug はSlugが一致するAppを返します。
func (s *MemoryAppStore) FindBySlug(ctx context.Context, slug domain.Slug) (*domain.App, bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}

	for i := range s.apps {
		if s.apps[i].Slug() == slug {
			app := s.apps[i]
			return &app, true, nil
		}
	}
	return nil, false, nil
}
