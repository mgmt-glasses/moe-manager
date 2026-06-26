package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mgmt-glasses/moe-manager/internal/auth"
)

type fakeVerifier struct {
	user auth.User
	err  error
}

func (v fakeVerifier) VerifyIDToken(_ context.Context, _ string) (auth.User, error) {
	return v.user, v.err
}

func TestRequireAuthStoresVerifiedUser(t *testing.T) {
	mw := auth.NewMiddleware(fakeVerifier{user: auth.User{UID: "u_001", Email: "u@example.com"}})
	called := false

	handler := mw.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		user, ok := auth.UserFromContext(r.Context())
		if !ok || user.UID != "u_001" {
			t.Fatalf("expected authenticated user, got %#v ok=%v", user, ok)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected next handler to be called")
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestRequireAuthRejectsMissingToken(t *testing.T) {
	mw := auth.NewMiddleware(fakeVerifier{user: auth.User{UID: "u_001"}})
	handler := mw.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequirePathUserRejectsOtherUser(t *testing.T) {
	mw := auth.NewMiddleware(fakeVerifier{user: auth.User{UID: "u_001"}})
	r := chi.NewRouter()
	r.Use(mw.RequireAuth)
	r.Route("/users/{userId}", func(r chi.Router) {
		r.Use(mw.RequirePathUser("userId"))
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("next handler should not be called")
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/users/u_002/", nil)
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}
