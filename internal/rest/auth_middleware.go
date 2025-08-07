package rest

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/elangreza/content-management-system/internal/constanta"
	"github.com/google/uuid"
)

type (
	authService interface {
		ProcessToken(ctx context.Context, reqToken string) (uuid.UUID, error)
	}

	AuthMiddleware struct {
		authService
	}
)

func NewAuthMiddleware(AuthService authService) *AuthMiddleware {
	return &AuthMiddleware{AuthService}
}

func (am *AuthMiddleware) MustAuthMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			rawAuthorization := r.Header["Authorization"]
			if len(rawAuthorization) == 0 {
				sendErrorResponse(w, http.StatusBadRequest, errors.New("token not valid"))
				return
			}

			authorization := rawAuthorization[0]

			rawToken := strings.Split(authorization, " ")
			if len(rawToken) != 2 {
				sendErrorResponse(w, http.StatusBadRequest, errors.New("token not valid"))
				return
			}

			token := rawToken[1]

			userID, err := am.authService.ProcessToken(r.Context(), token)
			if err != nil {
				sendErrorResponse(w, http.StatusUnauthorized, errors.New("cannot unauthorize this user"))
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), constanta.LocalUserID, userID.String()))

			next.ServeHTTP(w, r)
		})
	}
}
