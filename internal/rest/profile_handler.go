package rest

import (
	"context"
	"net/http"

	"github.com/elangreza/content-management-system/internal/params"
)

type (
	ProfileService interface {
		GetUserProfile(ctx context.Context) (*params.UserProfileResponse, error)
	}

	ProfileHandler struct {
		svc ProfileService
	}
)

// NewProfileHandler initializes the ProfileHandler with the given service.
//
//	@Summary		Get User Profile
//	@Description	Get the profile of the authenticated user.
//	@Tags			Profile
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string	false	"fill with Bearer token for accessing draft, published, and archived articles. If authorization is not provided default is showing published articles"
//	@Success		200				{object}	params.UserProfileResponse
//	@Failure		500				{object}	APIError
//	@Router			/profile [get]
func (ah *ProfileHandler) ProfileUserHandler(w http.ResponseWriter, r *http.Request) {

	profile, err := ah.svc.GetUserProfile(r.Context())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	sendSuccessResponse(w, http.StatusOK, profile)
}
