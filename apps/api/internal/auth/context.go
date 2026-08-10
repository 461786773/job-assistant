package auth

import "context"

type ctxKey int

const userKey ctxKey = 1

func WithUser(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, userKey, claims)
}

func UserFromContext(ctx context.Context) *Claims {
	v, _ := ctx.Value(userKey).(*Claims)
	return v
}
