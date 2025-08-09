package params

import (
	"strings"
	"time"

	"github.com/elangreza/content-management-system/internal/constanta"
	errs "github.com/elangreza/content-management-system/internal/error"
	"github.com/google/uuid"
)

type CreateArticleRequest struct {
	Title, Body string
	Tags        []string
}

type CreateArticleResponse struct {
	ArticleID int64
}

type UpdateArticleStatusRequest struct {
	Status int8 `json:"status"`
}

type CreateArticleVersionRequest struct {
	Title, Body string
	Tags        []string
}

type CreateArticleVersionResponse struct {
	ArticleVersionID int64
}

type ArticleVersionResponse struct {
	ArticleID int64
	VersionID int64
	Title     string
	Body      string
	Version   int64
	Status    int8

	CreatedBy uuid.UUID
	CreatedAt time.Time
	UpdatedBy uuid.UUID
	UpdatedAt *time.Time
}

type GetArticleDetailResponse struct {
	ID               int64                   `json:"id"`
	DraftedVersion   *ArticleVersionResponse `json:"drafted_version"`
	PublishedVersion *ArticleVersionResponse `json:"published_version"`

	CreatedBy uuid.UUID
	CreatedAt time.Time
	UpdatedBy uuid.UUID
	UpdatedAt time.Time
}

type GetArticlesQueryParams struct {
	// can be searched by title, content
	Search    string
	Status    []constanta.ArticleVersionStatus
	CreatedBy []uuid.UUID
	UpdatedBy []uuid.UUID
	// TODO tags

	// Embedding PaginationParams for pagination and sorting
	PaginationParams
}

func (pqr *GetArticlesQueryParams) Validate() error {

	if len(pqr.Sorts) == 0 {
		pqr.Sorts = append(pqr.Sorts, "created_at:desc")
	}

	for _, v := range pqr.Status {
		if v < constanta.Draft || v > constanta.Archived {
			return errs.ValidationError{Message: "not valid status"}
		}
	}

	if len(pqr.Status) == 0 {
		pqr.Status = []constanta.ArticleVersionStatus{
			constanta.Published,
		}
	}

	pqr.PaginationParams.setValidSortKey(
		"article_id",
		"article_version_id",
		"crated_by",
		"updated_by",
		"title",
		"status",
		"version",
		"created_at",
		"updated_at",
		// TODO implement this
		"tag_relationship_score",
	)

	if err := pqr.PaginationParams.Validate(); err != nil {
		return errs.ValidationError{Message: err.Error()}
	}

	pqr.Search = strings.TrimSpace(pqr.Search)

	return nil
}
