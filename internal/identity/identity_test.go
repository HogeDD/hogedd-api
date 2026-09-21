package identity

import (
	"context"
	"testing"
)

func TestIdentityRequiresIssuerAndSubject(t *testing.T) {
	tests := []struct {
		name    string
		issuer  string
		subject string
		wantErr bool
	}{
		{name: "valid", issuer: "https://hogedd.jp.auth0.com/", subject: "auth0|owner"},
		{name: "missing issuer", subject: "auth0|owner", wantErr: true},
		{name: "missing subject", issuer: "https://hogedd.jp.auth0.com/", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.issuer, tt.subject)
			if (err != nil) != tt.wantErr {
				t.Fatalf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && (got.Issuer() != tt.issuer || got.Subject() != tt.subject) {
				t.Errorf("Identity = (%q, %q)", got.Issuer(), got.Subject())
			}
		})
	}
}

func TestIdentityContextRoundTrip(t *testing.T) {
	authenticated, err := New("https://hogedd.jp.auth0.com/", "auth0|owner")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx := NewContext(context.Background(), authenticated)
	got, ok := FromContext(ctx)
	if !ok {
		t.Fatal("FromContext() found no Identity")
	}
	if got != authenticated {
		t.Errorf("FromContext() = %+v, want %+v", got, authenticated)
	}
}
