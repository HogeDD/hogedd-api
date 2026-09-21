package identity

import (
	"context"
	"fmt"
)

// Identity は信頼できる検証器が確定した外部の認証主体です。
// IssuerとSubjectの組み合わせをHogeDD内部で一意な外部identityとして扱います。
type Identity struct {
	issuer  string
	subject string
}

// New は検証済みclaimからIdentityを構築します。
func New(issuer, subject string) (Identity, error) {
	if issuer == "" {
		return Identity{}, fmt.Errorf("identity issuer is required")
	}
	if subject == "" {
		return Identity{}, fmt.Errorf("identity subject is required")
	}
	return Identity{issuer: issuer, subject: subject}, nil
}

// Issuer は認証主体を発行した認証基盤の識別子を返します。
func (i Identity) Issuer() string {
	return i.issuer
}

// Subject はissuer内で認証主体を一意に識別する値を返します。
func (i Identity) Subject() string {
	return i.subject
}

type contextKey struct{}

// NewContext は検証済みIdentityをrequest scopeへ保存したcontextを返します。
func NewContext(ctx context.Context, authenticated Identity) context.Context {
	return context.WithValue(ctx, contextKey{}, authenticated)
}

// FromContext はrequest scopeに保存された検証済みIdentityを返します。
func FromContext(ctx context.Context) (Identity, bool) {
	authenticated, ok := ctx.Value(contextKey{}).(Identity)
	return authenticated, ok
}
