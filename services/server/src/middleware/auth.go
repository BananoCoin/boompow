package middleware

import (
	"context"
	"net/http"

	"github.com/bananocoin/boompow-next/libs/utils/auth"
	"github.com/bananocoin/boompow-next/services/server/src/models"
	"github.com/bananocoin/boompow-next/services/server/src/repository"
)

var userCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

func Middleware(userRepo *repository.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")

			// Allow unauthenticated users in
			if header == "" {
				next.ServeHTTP(w, r)
				return
			}

			//validate jwt token
			tokenStr := header
			email, err := auth.ParseToken(tokenStr)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusForbidden)
				return
			}

			// create user and check if user exists in db
			user, err := userRepo.GetUser(nil, &email)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			// put it in context
			ctx := context.WithValue(r.Context(), userCtxKey, user)

			// and call the next with our new context
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// ForContext finds the user from the context. REQUIRES Middleware to have run.
func ForContext(ctx context.Context) *models.User {
	raw, _ := ctx.Value(userCtxKey).(*models.User)
	return raw
}
