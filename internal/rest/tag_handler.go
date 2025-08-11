package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	errs "github.com/elangreza/content-management-system/internal/error"
	"github.com/elangreza/content-management-system/internal/params"
	"github.com/go-chi/chi/v5"
)

type (
	TagService interface {
		CreateTag(ctx context.Context, tagNames ...string) error
		GetTags(ctx context.Context, req params.GetTagsRequest) ([]params.GetTagResponse, error)
		GetTag(ctx context.Context, tagName string) (*params.GetTagResponse, error)
	}

	TagHandler struct {
		svc TagService
	}
)

// CreateTagHandler creates a new tag.
//
//	@Summary		Create Tag
//	@Description	Create a new tag with the provided names.
//	@Tags			Tags
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"MUST HAVE PERMISSION CreateArticle. Fill with Bearer token. The token can be accessed via api /auth/login."
//	@Param			body			body		params.CreateTagRequest	true	"Create Tag Request"
//	@Success		201				{string}	string					"created"
//	@Failure		400				{object}	errs.ValidationError
//	@Failure		500				{object}	APIError
//	@Router			/tags [post]
func (ah *TagHandler) CreateTagHandler(w http.ResponseWriter, r *http.Request) {

	body := params.CreateTagRequest{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, errs.ValidationError{Message: err.Error()})
		return
	}

	if err := body.Validate(); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	err := ah.svc.CreateTag(r.Context(), body.Names...)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	sendSuccessResponse(w, http.StatusCreated, "created")
}

// GetTagsHandler retrieves a list of tags with optional sorting.
//
//	@Summary		Get Tags
//	@Description	Retrieve a list of tags with optional sorting.
//	@Tags			Tags
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string	false	"fill with Bearer token. The token can be accessed via api /auth/login. If authorization is not provided, the default behavior is showing only published articles. Otherwise if the token is appered and user habe a permission to read drafted and archiver article, the token can be used to accessing draft, published, and archived articles. "
//	@Param			sort			query		string	false	"Sort by usage_count:asc | usage_count:desc | trending_score:asc | trending_score:desc | name:asc | name:desc | last_used:asc | last_used:desc"
//	@Success		200				{array}		params.GetTagResponse
//	@Failure		400				{object}	errs.ValidationError
//	@Failure		500				{object}	APIError
//	@Router			/tags [get]
func (ah *TagHandler) GetTagsHandler(w http.ResponseWriter, r *http.Request) {

	sort := r.URL.Query().Get("sort")
	if sort == "" {
		sort = "usage_count:desc"
	}

	sortValue := strings.Split(sort, ":")
	if len(sortValue) != 2 {
		sendErrorResponse(w, http.StatusBadRequest, errs.ValidationError{Message: "invalid sort"})
		return
	}

	allowedSorts := map[string]struct{}{"usage_count": {}, "trending_score": {}, "name": {}, "last_used": {}}
	if _, ok := allowedSorts[sortValue[0]]; !ok {
		sendErrorResponse(w, http.StatusBadRequest, errs.ValidationError{Message: "invalid sort"})
		return
	}

	if len(sortValue) > 1 && sortValue[1] != "asc" && sortValue[1] != "desc" {
		sendErrorResponse(w, http.StatusBadRequest, errs.ValidationError{Message: "invalid sort order"})
		return
	}

	req := params.GetTagsRequest{
		SortValue: sortValue[0],
		Direction: sortValue[1],
	}

	tags, err := ah.svc.GetTags(r.Context(), req)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	sendSuccessResponse(w, http.StatusOK, tags)
}

// GetTagHandler retrieves a specific tag by name.
//
//	@Summary		Get Tag
//	@Description	Retrieve a specific tag by name.
//	@Tags			Tags
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string	false	"fill with Bearer token. The token can be accessed via api /auth/login. If authorization is not provided, the default behavior is showing only published articles. Otherwise if the token is appered and user habe a permission to read drafted and archiver article, the token can be used to accessing draft, published, and archived articles. "
//	@Param			name			path		string	true	"Tag name"
//	@Success		200				{object}	params.GetTagResponse
//	@Failure		400				{object}	errs.ValidationError
//	@Failure		500				{object}	APIError
//	@Router			/tags/{name} [get]
func (ah *TagHandler) GetTagHandler(w http.ResponseWriter, r *http.Request) {
	tagName := chi.URLParam(r, "name")
	if tagName == "" {
		sendErrorResponse(w, http.StatusBadRequest, errs.ValidationError{Message: "tag name is required"})
		return
	}
	name := strings.TrimSpace(tagName)
	if name == "" {
		sendErrorResponse(w, http.StatusBadRequest, errs.ValidationError{Message: "tag name cannot be empty"})
		return
	}

	tags, err := ah.svc.GetTag(r.Context(), name)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	sendSuccessResponse(w, http.StatusOK, tags)
}
