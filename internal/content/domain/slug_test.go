package domain

import (
	"errors"
	"testing"
)

func TestNewSlug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "英小文字", value: "nishida"},
		{name: "数字", value: "app2"},
		{name: "ハイフン区切り", value: "clean-tasks"},
		{name: "空文字", value: "", wantErr: true},
		{name: "英大文字", value: "Clean-Tasks", wantErr: true},
		{name: "先頭のハイフン", value: "-clean-tasks", wantErr: true},
		{name: "末尾のハイフン", value: "clean-tasks-", wantErr: true},
		{name: "連続するハイフン", value: "clean--tasks", wantErr: true},
		{name: "スラッシュ", value: "apps/clean-tasks", wantErr: true},
		{name: "日本語", value: "ニシ打", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewSlug(tt.value)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidSlug) {
					t.Fatalf("NewSlug() error = %v, want %v", err, ErrInvalidSlug)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewSlug() unexpected error = %v", err)
			}
			if got.String() != tt.value {
				t.Fatalf("Slug.String() = %q, want %q", got.String(), tt.value)
			}
		})
	}
}
