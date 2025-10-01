package middleware

import (
	"common/utils/auth"
	"context"
	"errors"
	"net/http"
	"strings"
)

type ctxKey string

const (
	CtxUserID ctxKey = "userID"
	CtxRoles  ctxKey = "roles"
)

func Authenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz := r.Header.Get("Authorization")
		if !strings.HasPrefix(authz, "Bearer ") {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(authz, "Bearer ")

		claims, err := auth.ValidateToken(token) // common util, ensures typ=access
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		userId, _ := claims["sub"].(string)

		ctx := context.WithValue(r.Context(), CtxUserID, userId)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserID(ctx context.Context) (string, error) {
	v := ctx.Value(CtxUserID)
	s, _ := v.(string)
	if s == "" {
		return "", errors.New("no user in context")
	}
	return s, nil
}
