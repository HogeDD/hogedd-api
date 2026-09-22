package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestNewProfileNormalizesDisplayName(t *testing.T) {
	profile, err := NewProfile("user-1", "  HogeDD\tOwner  ")
	if err != nil {
		t.Fatal(err)
	}
	if profile.DisplayName() != "HogeDD Owner" {
		t.Fatalf("DisplayName() = %q", profile.DisplayName())
	}
}

func TestNewProfileRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		displayName string
		want        error
	}{
		{name: "missing user", displayName: "name", want: ErrUserIDRequired},
		{name: "empty name", userID: "user-1", displayName: " \t ", want: ErrDisplayNameRequired},
		{name: "long name", userID: "user-1", displayName: strings.Repeat("あ", MaxDisplayNameLength+1), want: ErrDisplayNameTooLong},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewProfile(tt.userID, tt.displayName)
			if !errors.Is(err, tt.want) {
				t.Fatalf("NewProfile() error = %v, want %v", err, tt.want)
			}
		})
	}
}
