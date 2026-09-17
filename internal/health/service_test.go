package health

import (
	"context"
	"testing"
)

func TestServiceLiveness(t *testing.T) {
	t.Parallel()

	got := NewService().Liveness(context.Background())
	if got.Status != "ok" {
		t.Fatalf("Status = %q, want %q", got.Status, "ok")
	}
}
