package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/mgmt-glasses/moe-manager/internal/shared"
)

var (
	ErrMissingToken = errors.New("missing bearer token")
	ErrInvalidToken = errors.New("invalid bearer token")
)

type TokenVerifier interface {
	VerifyIDToken(ctx context.Context, token string) (User, error)
}

type Middleware struct {
	verifier TokenVerifier
}

func NewMiddleware(verifier TokenVerifier) *Middleware {
	return &Middleware{verifier: verifier}
}

func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := bearerToken(r.Header.Get("Authorization"))
		if err != nil {
			shared.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "認証が必要です")
			return
		}

		user, err := m.verifier.VerifyIDToken(r.Context(), token)
		if err != nil {
			shared.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "認証トークンが正しくありません")
			return
		}
		if user.UID == "" {
			shared.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "認証トークンが正しくありません")
			return
		}

		next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), user)))
	})
}

func (m *Middleware) RequirePathUser(paramName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok || user.UID == "" {
				shared.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "認証が必要です")
				return
			}
			if chi.URLParam(r, paramName) != user.UID {
				shared.WriteError(w, http.StatusForbidden, "FORBIDDEN", "他ユーザーのデータにはアクセスできません")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func bearerToken(header string) (string, error) {
	if header == "" {
		return "", ErrMissingToken
	}
	typ, token, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(typ, "Bearer") || strings.TrimSpace(token) == "" {
		return "", ErrInvalidToken
	}
	return strings.TrimSpace(token), nil
}
