package api

import "context"

type contextKey string

const ctxUserID = contextKey("user_id")

func contextWithUserID(ctx context.Context, userID uint) context.Context {
	return context.WithValue(ctx, ctxUserID, userID)
}

func userIDFromContext(ctx context.Context) (uint, bool) {
	userID, ok := ctx.Value(ctxUserID).(uint)
	return userID, ok
}
