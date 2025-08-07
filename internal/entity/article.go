package entity

import (
	"time"

	"github.com/google/uuid"
)

type ArticleVersionStatus int8

const (
	Pending ArticleVersionStatus = iota
	Published
	Archived
)

type (
	Article struct {
		ID                 int64
		PublishedVersionID int64
		DraftedVersionID   int64
		Versions           []ArticleVersion

		CreatedBy uuid.UUID
		CreatedAt time.Time
		UpdatedBy uuid.UUID
		UpdatedAt time.Time
	}

	ArticleVersion struct {
		ID        int64
		ArticleID int64
		Title     string
		Body      string
		Version   int64
		Status    ArticleVersionStatus

		CreatedBy uuid.UUID
		CreatedAt time.Time
		UpdatedBy uuid.UUID
		UpdatedAt time.Time
	}
)

func NewArticle(title, body string, createdBy uuid.UUID) *Article {
	return &Article{
		CreatedBy: createdBy,
		Versions: []ArticleVersion{
			{
				Title:     title,
				Body:      body,
				Status:    Pending,
				Version:   1,
				CreatedBy: createdBy,
			},
		},
	}
}
