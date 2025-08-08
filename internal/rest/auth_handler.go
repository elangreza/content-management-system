package rest

import (
	"context"
	"encoding/json"
	"net/http"

	errs "github.com/elangreza/content-management-system/internal/error"
	"github.com/elangreza/content-management-system/internal/params"
	"github.com/go-chi/chi/v5"
)

type (
	AutService interface {
		RegisterUser(ctx context.Context, req params.RegisterUserRequest) error
		LoginUser(ctx context.Context, req params.LoginUserRequest) (string, error)
	}

	AuthHandler struct {
		svc AutService
	}
)

func NewAuthRouter(ar chi.Router, authService AutService) {

	authHandler := AuthHandler{
		svc: authService,
	}

	ar.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.RegisterUser)
		r.Post("/login", authHandler.LoginUser)
	})
}

func (ah *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	body := params.RegisterUserRequest{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, errs.ValidationError{Message: err.Error()})
		return
	}

	// TODO validation
	// if err := body.Validate(); err != nil {
	// 	Error(w, http.StatusBadRequest, err)
	// 	return
	// }

	err := ah.svc.RegisterUser(r.Context(), body)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	sendSuccessResponse(w, http.StatusCreated, "ok")
}

func (ah *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	body := params.LoginUserRequest{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, errs.ValidationError{Message: err.Error()})
		return
	}

	// TODO validation
	// if err := body.Validate(); err != nil {
	// 	Error(w, http.StatusBadRequest, err)
	// 	return
	// }

	res, err := ah.svc.LoginUser(r.Context(), body)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	sendSuccessResponse(w, http.StatusOK, res)
}
