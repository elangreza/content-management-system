package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	errs "github.com/elangreza/content-management-system/internal/error"
	"github.com/elangreza/content-management-system/internal/params"
)

type (
	TagService interface {
		CreateTag(ctx context.Context, tagNames ...string) error
		GetTags(ctx context.Context, req params.GetTagsRequest) ([]params.GetTagResponse, error)
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

	sort := r.URL.Query().Get("sort")
	if sort == "" {
		sort = "usage_count:desc"
	}

	sortValue := strings.Split(sort, ":")
	if len(sortValue) != 2 {
		sendErrorResponse(w, http.StatusBadRequest, errs.ValidationError{Message: "invalid sort"})
		return
	}

	allowedSorts := map[string]struct{}{"usage_count": {}, "trending_score": {}, "name": {}}
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
