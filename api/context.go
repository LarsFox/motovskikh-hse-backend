package api

import (
	"net/http"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

//nolint:unused
type contextKey int

//nolint:unused
const ctxUser contextKey = iota + 1

//nolint:unused
func getFromCtxTester(r *http.Request) *entities.User {
	tester, ok := r.Context().Value(ctxUser).(*entities.User)
	if !ok {
		return nil
	}
	return tester
}
