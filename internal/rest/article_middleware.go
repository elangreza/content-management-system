package rest

import (
	"context"
	"errors"
	"net/http"

	"github.com/elangreza/content-management-system/internal/constanta"
	"github.com/google/uuid"
)

type (
	ArticleMiddleware struct {
		svc AuthService
	}
)

func (am *ArticleMiddleware) CanSeeDraftOrArchivedArticle() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ctx := r.Context()

			var canSeeDraftOrArchivedArticle bool
			userID, ok := ctx.Value(constanta.LocalUserID).(uuid.UUID)
			if ok {
				userRole, err := am.svc.GetUserRoleByUserID(ctx, userID)
				if err != nil {
					sendErrorResponse(w, http.StatusInternalServerError, errors.New("error when get user role"))
					return
				}

				if userRole.HasPermission(constanta.ReadDraftedAndArchivedArticle) {
					canSeeDraftOrArchivedArticle = true
				}

				ctx = context.WithValue(ctx, constanta.LocalUserRole, userRole.GetValue())
			}

			ctx = context.WithValue(ctx, constanta.LocalUserCanReadDraftedAndArchivedArticle, canSeeDraftOrArchivedArticle)

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
