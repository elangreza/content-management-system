package params

import (
	"time"

	"github.com/google/uuid"
)

type CreateArticleRequest struct {
	Title, Body string
}

type CreateArticleResponse struct {
	ArticleID int64
}

type UpdateArticleStatusRequest struct {
	Status int8 `json:"status"`
}

type CreateArticleVersionRequest struct {
	Title, Body string
}

type CreateArticleVersionResponse struct {
	ArticleVersionID int64
}

type ArticleVersionResponse struct {
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
