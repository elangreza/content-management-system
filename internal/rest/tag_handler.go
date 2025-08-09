package rest

import (
	"context"
	"encoding/json"
	"net/http"

	errs "github.com/elangreza/content-management-system/internal/error"
	"github.com/elangreza/content-management-system/internal/params"
)

type (
	TagService interface {
		CreateTag(ctx context.Context, tagNames ...string) error
		GetTags(ctx context.Context) ([]params.GetTagResponse, error)
	}

	TagHandler struct {
		svc TagService
	}
)

func (ah *TagHandler) CreateTagHandler(w http.ResponseWriter, r *http.Request) {

	body := params.CreateTagRequest{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, errs.ValidationError{Message: err.Error()})
		return
	}

	// TODO validation
	// if err := body.Validate(); err != nil {
	// 	Error(w, http.StatusBadRequest, err)
	// 	return
	// }

	err := ah.svc.CreateTag(r.Context(), body.Names...)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	sendSuccessResponse(w, http.StatusCreated, "created")
}

func (ah *TagHandler) GetTagsHandler(w http.ResponseWriter, r *http.Request) {
	tags, err := ah.svc.GetTags(r.Context())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	sendSuccessResponse(w, http.StatusOK, tags)
}
