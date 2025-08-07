package rest

import (
	"context"

	"github.com/elangreza/content-management-system/internal/params"
	"github.com/google/uuid"
)

type (
	authServiceReq interface {
		RegisterUser(ctx context.Context, req params.RegisterUserRequest) error
		LoginUser(ctx context.Context, req params.LoginUserRequest) (string, error)
		ProcessToken(ctx context.Context, reqToken string) (uuid.UUID, error)
	}

	profileService interface {
		GetUserProfile(ctx context.Context) (*params.UserProfileResponse, error)
	}
)

// func NewRoute(as authServiceReq, ps profileService) {

// 	// ar := chi.NewRouter()

// 	// am := NewAuthMiddleware(as)

// 	pr := NewProfileRouter(ps)
// 	pr.Use(am.MustAuthMiddleware())

// }
