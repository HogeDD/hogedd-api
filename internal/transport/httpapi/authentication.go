package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/iwasawa/hogedd-api/internal/identity"
)

// AccessTokenVerifier はBearer Tokenを検証済みIdentityへ変換します。
// 実装は署名、issuer、audience、有効期限をすべて検証する必要があります。
type AccessTokenVerifier interface {
	Verify(context.Context, string) (identity.Identity, error)
}

// AuthenticateBearer は有効なBearer Tokenを要求し、検証済みIdentityをcontextへ保存します。
func AuthenticateBearer(verifier AccessTokenVerifier, responder *Responder) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r.Header.Values("Authorization"))
			if !ok {
				unauthorized(w, r, responder)
				return
			}

			authenticated, err := verifier.Verify(r.Context(), token)
			if err != nil {
				unauthorized(w, r, responder)
				return
			}

			ctx := identity.NewContext(r.Context(), authenticated)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(values []string) (string, bool) {
	if len(values) != 1 {
		return "", false
	}
	parts := strings.Fields(values[0])
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	return parts[1], true
}

func unauthorized(w http.ResponseWriter, r *http.Request, responder *Responder) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	responder.Error(w, r, http.StatusUnauthorized, "unauthorized", "valid access token is required")
}
