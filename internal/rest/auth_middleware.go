package rest

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/elangreza/content-management-system/internal/constanta"
	"github.com/elangreza/content-management-system/internal/entity"
	"github.com/google/uuid"
)

type (
	AuthService interface {
		ProcessToken(ctx context.Context, reqToken string) (uuid.UUID, error)
		GetUserRoleByUserID(ctx context.Context, id uuid.UUID) (*entity.UserRole, error)
	}

	AuthMiddleware struct {
		svc AuthService
	}
)

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

			userID, err := am.svc.ProcessToken(r.Context(), token)
			if err != nil {
				sendErrorResponse(w, http.StatusUnauthorized, errors.New("cannot unauthorize this user"))
				return
			}

			ctx := context.WithValue(r.Context(), constanta.LocalUserID, userID)

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

func (am *AuthMiddleware) MustHavePermission(permissions ...constanta.UserPermission) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			userID, ok := ctx.Value(constanta.LocalUserID).(uuid.UUID)
			if !ok {
				sendErrorResponse(w, http.StatusBadRequest, errors.New("error when handle ctx value"))
				return
			}

			userRole, err := am.svc.GetUserRoleByUserID(ctx, userID)
			if err != nil {
				sendErrorResponse(w, http.StatusInternalServerError, errors.New("error when get user role"))
				return
			}

			for _, permission := range permissions {
				if !userRole.HasPermission(permission) {
					sendErrorResponse(w, http.StatusForbidden, errors.New("you dont have permission to access this resource"))
					return
				}
			}

			ctx = context.WithValue(ctx, constanta.LocalUserRole, userRole.GetValue())

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

func (am *AuthMiddleware) OptionalAuthMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			rawAuthorization := r.Header["Authorization"]
			if len(rawAuthorization) > 0 {
				authorization := rawAuthorization[0]

				rawToken := strings.Split(authorization, " ")
				if len(rawToken) != 2 {
					sendErrorResponse(w, http.StatusBadRequest, errors.New("token not valid"))
					return
				}

				token := rawToken[1]

				userID, err := am.svc.ProcessToken(r.Context(), token)
				if err != nil {
					sendErrorResponse(w, http.StatusUnauthorized, errors.New("cannot unauthorize this user"))
					return
				}

				ctx := context.WithValue(r.Context(), constanta.LocalUserID, userID)

				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}
