package domain

import "testing"

func TestPublicationStatusIsPublic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status PublicationStatus
		want   bool
	}{
		{name: "準備中", status: PublicationStatusPreparing, want: false},
		{name: "公開済み", status: PublicationStatusPublished, want: true},
		{name: "未定義の状態", status: PublicationStatus("unknown"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.status.IsPublic(); got != tt.want {
				t.Fatalf("PublicationStatus.IsPublic() = %t, want %t", got, tt.want)
			}
		})
	}
}
