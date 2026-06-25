package auth

import "context"

type contextKey struct{}

// User is the authenticated caller resolved from a verified ID token.
type User struct {
	UID   string
	Email string
}

func WithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, contextKey{}, user)
}

func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(contextKey{}).(User)
	return user, ok
}
