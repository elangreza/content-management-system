package entity

import (
	"time"

	"github.com/elangreza/content-management-system/internal/constanta"
	"github.com/google/uuid"
)

type (
	Article struct {
		ID                 int64
		PublishedVersionID int64
		DraftedVersionID   int64
		VersionSequence    int64
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
		Status    constanta.ArticleVersionStatus

		CreatedBy uuid.UUID
		CreatedAt time.Time
		UpdatedBy uuid.UUID
		UpdatedAt *time.Time
	}
)

func NewArticle(title, body string, createdBy uuid.UUID) *Article {
	return &Article{
		CreatedBy:       createdBy,
		VersionSequence: 1,
		Versions: []ArticleVersion{
			{
				Title:     title,
				Body:      body,
				Status:    constanta.Draft,
				Version:   1,
				CreatedBy: createdBy,
			},
		},
	}
}

func NewArticleVersion(articleID int64, title, body string, createdBy uuid.UUID, version int64) *ArticleVersion {
	return &ArticleVersion{
		ArticleID: articleID,
		Title:     title,
		Body:      body,
		Status:    constanta.Draft,
		Version:   version,
		CreatedBy: createdBy,
	}
}
