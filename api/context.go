package api

import (
	"net/http"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

type contextKey int

const ctxUser contextKey = iota + 1

func getFromCtxTester(r *http.Request) *entities.User {
	tester, ok := r.Context().Value(ctxUser).(*entities.User)
	if !ok {
		return nil
	}
	return tester
}
