package rest

import (
	"context"
	"net/http"

	"github.com/elangreza/content-management-system/internal/params"
	"github.com/go-chi/chi/v5"
)

type (
	ProfileService interface {
		GetUserProfile(ctx context.Context) (*params.UserProfileResponse, error)
	}

	ProfileHandler struct {
		svc ProfileService
	}
)

func NewProfileRouter(ar chi.Router, profileService ProfileService) {

	profileHandler := ProfileHandler{
		svc: profileService,
	}

	ar.Get("/profile", profileHandler.ProfileUserHandler)
}

func (ah *ProfileHandler) ProfileUserHandler(w http.ResponseWriter, r *http.Request) {

	profile, err := ah.svc.GetUserProfile(r.Context())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	sendSuccessResponse(w, http.StatusOK, profile)
}
