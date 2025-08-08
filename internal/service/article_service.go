package service

import (
	"context"
	"errors"

	"github.com/elangreza/content-management-system/internal/constanta"
	"github.com/elangreza/content-management-system/internal/entity"
	errs "github.com/elangreza/content-management-system/internal/error"
	"github.com/elangreza/content-management-system/internal/params"
	"github.com/google/uuid"
)

type (
	articleRepo interface {
		CreateArticle(ctx context.Context, article entity.Article) (int64, error)
		DeleteArticle(ctx context.Context, articleID int64) error
		GetArticleVersionWithIDAndArticleID(ctx context.Context, articleID int64, articleVersionID int64) (*entity.ArticleVersion, error)
		UpdateArticleVersion(ctx context.Context, articleID int64, articleVersionID int64, status constanta.ArticleVersionStatus, updatedBy uuid.UUID) error
		CreateArticleVersion(ctx context.Context, articleVersion entity.ArticleVersion) (int64, error)
		GetArticleWithID(ctx context.Context, articleID int64) (*entity.Article, error)
		GetArticleVersionsWithArticleIDAndStatuses(ctx context.Context, ArticleID int64, status ...constanta.ArticleVersionStatus) ([]entity.ArticleVersion, error)
	}

	ArticleService struct {
		ArticleRepo articleRepo
	}
)

func NewArticleService(articleRepo articleRepo) *ArticleService {
	return &ArticleService{
		ArticleRepo: articleRepo,
	}
}

// => POST /articles
func (as *ArticleService) CreateArticle(ctx context.Context, req params.CreateArticleRequest) (*params.CreateArticleResponse, error) {
	userID, ok := ctx.Value(constanta.LocalUserID).(uuid.UUID)
	if !ok {
		return nil, errors.New("error when parsing userID")
	}

	article := entity.NewArticle(req.Title, req.Body, userID)
	id, err := as.ArticleRepo.CreateArticle(ctx, *article)
	if err != nil {
		return nil, err
	}

	return &params.CreateArticleResponse{
		ArticleID: id,
	}, nil
}

// => DELETE /articles/{id}
func (as *ArticleService) DeleteArticle(ctx context.Context, articleID int64) error {
	return as.ArticleRepo.DeleteArticle(ctx, articleID)
}

// => PUT /articles/{id}/versions/{id}/status
func (as *ArticleService) UpdateStatusArticle(ctx context.Context, articleID, articleVersionID int64, reqStatus constanta.ArticleVersionStatus) error {
	userID, ok := ctx.Value(constanta.LocalUserID).(uuid.UUID)
	if !ok {
		return errors.New("error when parsing userID")
	}

	articleVersion, err := as.ArticleRepo.GetArticleVersionWithIDAndArticleID(ctx, articleID, articleVersionID)
	if err != nil {
		return err
	}

	if reqStatus == articleVersion.Status {
		return errs.ValidationError{Message: "status cannot be same as current status"}
	}

	if reqStatus < articleVersion.Status {
		return errs.ValidationError{Message: "status cannot be downgraded"}
	}

	return as.ArticleRepo.UpdateArticleVersion(ctx, articleID, articleVersionID, reqStatus, userID)
}

// => POST /articles/{id}/versions/{id}
func (as *ArticleService) CreateArticleVersion(ctx context.Context, articleID int64, articleVersionID int64, req params.CreateArticleVersionRequest) (*params.CreateArticleVersionResponse, error) {
	userID, ok := ctx.Value(constanta.LocalUserID).(uuid.UUID)
	if !ok {
		return nil, errors.New("error when parsing userID")
	}

	articleVersion, err := as.ArticleRepo.GetArticleVersionWithIDAndArticleID(ctx, articleID, articleVersionID)
	if err != nil {
		return nil, err
	}

	if articleVersion.Title == req.Title && articleVersion.Body == req.Body {
		return nil, errs.ValidationError{Message: "title and body cannot be the same as the current version"}
	}

	article, err := as.ArticleRepo.GetArticleWithID(ctx, articleID)
	if err != nil {
		return nil, err
	}

	version := article.VersionSequence + 1
	newArticleVersion := entity.NewArticleVersion(articleID, req.Title, req.Body, userID, version)
	newArticleVersionID, err := as.ArticleRepo.CreateArticleVersion(ctx, *newArticleVersion)
	if err != nil {
		return nil, err
	}

	return &params.CreateArticleVersionResponse{
		ArticleVersionID: newArticleVersionID,
	}, nil
}

// DONE
// Pembuatan Artikel Baru
// => POST /articles
// Penghapusan Artikel
// => DELETE /articles/{id}
// Perubahan Status Versi Artikel
// => PUT /articles/{id}/versions/{id}/status
// Pembuatan Versi Artikel Baru
// => POST /articles/{id}/versions/{id}
// Pengambilan Detail Artikel Terbaru
// => POST /articles/{id}
// Pengambilan Detail Versi Artikel Tertentu
// => GET /articles/{id}/versions/{id}

// => POST /articles/{id}
func (as *ArticleService) GetArticleWithID(ctx context.Context, articleID int64) (*params.GetArticleDetailResponse, error) {
	article, err := as.ArticleRepo.GetArticleWithID(ctx, articleID)
	if err != nil {
		return nil, err
	}

	userCanReadDraftedAndArchivedArticle, ok := ctx.Value(constanta.LocalUserCanReadDraftedAndArchivedArticle).(bool)
	if !ok {
		return nil, errors.New("error when parsing user permission")
	}

	// TODO check permission ReadDraftedOrArchivedArticle

	var articleVersionResponse *params.ArticleVersionResponse
	if article.DraftedVersionID != 0 && userCanReadDraftedAndArchivedArticle {
		draftedVersion, err := as.ArticleRepo.GetArticleVersionWithIDAndArticleID(ctx, article.ID, article.DraftedVersionID)
		if err != nil {
			return nil, err
		}

		articleVersionResponse = &params.ArticleVersionResponse{
			VersionID: draftedVersion.ID,
			Title:     draftedVersion.Title,
			Body:      draftedVersion.Body,
			Version:   draftedVersion.Version,
			Status:    int8(draftedVersion.Status),
			CreatedBy: draftedVersion.CreatedBy,
			CreatedAt: draftedVersion.CreatedAt,
			UpdatedBy: draftedVersion.UpdatedBy,
			UpdatedAt: draftedVersion.UpdatedAt,
		}
	}

	var publishedVersionResponse *params.ArticleVersionResponse
	if article.PublishedVersionID != 0 {
		publishedVersion, err := as.ArticleRepo.GetArticleVersionWithIDAndArticleID(ctx, article.ID, article.PublishedVersionID)
		if err != nil {
			return nil, err
		}

		publishedVersionResponse = &params.ArticleVersionResponse{
			VersionID: publishedVersion.ID,
			Title:     publishedVersion.Title,
			Body:      publishedVersion.Body,
			Version:   publishedVersion.Version,
			Status:    int8(publishedVersion.Status),
			CreatedBy: publishedVersion.CreatedBy,
			CreatedAt: publishedVersion.CreatedAt,
			UpdatedBy: publishedVersion.UpdatedBy,
			UpdatedAt: publishedVersion.UpdatedAt,
		}
	}

	if articleVersionResponse == nil && publishedVersionResponse == nil {
		if !userCanReadDraftedAndArchivedArticle {
			return nil, errs.ValidationError{Message: "unauthenticated user cannot access this endpoint with status Drafted or Archived"}
		}

		return nil, errs.ValidationError{Message: "this article has no published or drafted version"}
	}

	return &params.GetArticleDetailResponse{
		ID:               article.ID,
		DraftedVersion:   articleVersionResponse,
		PublishedVersion: publishedVersionResponse,
		CreatedAt:        article.CreatedAt,
		CreatedBy:        article.CreatedBy,
		UpdatedAt:        article.UpdatedAt,
		UpdatedBy:        article.UpdatedBy,
	}, nil
}

// => GET /articles/{id}/versions/{id}
func (as *ArticleService) GetArticleVersionWithIDAndArticleID(ctx context.Context, articleID int64, articleVersionID int64) (*params.ArticleVersionResponse, error) {
	articleVersion, err := as.ArticleRepo.GetArticleVersionWithIDAndArticleID(ctx, articleID, articleVersionID)
	if err != nil {
		return nil, err
	}

	userCanReadDraftedAndArchivedArticle, ok := ctx.Value(constanta.LocalUserCanReadDraftedAndArchivedArticle).(bool)
	if !ok {
		return nil, errors.New("error when parsing user permission")
	}

	if !userCanReadDraftedAndArchivedArticle && (articleVersion.Status == constanta.Draft || articleVersion.Status == constanta.Archived) {
		return nil, errs.ValidationError{Message: "unauthenticated user cannot access this endpoint with status Drafted or Archived"}
	}

	return &params.ArticleVersionResponse{
		VersionID: articleVersion.ID,
		Title:     articleVersion.Title,
		Body:      articleVersion.Body,
		Version:   articleVersion.Version,
		Status:    int8(articleVersion.Status),
		CreatedBy: articleVersion.CreatedBy,
		CreatedAt: articleVersion.CreatedAt,
		UpdatedBy: articleVersion.UpdatedBy,
		UpdatedAt: articleVersion.UpdatedAt,
	}, nil
}

// TODO Pengambilan Daftar Versi Artikel
// => GET /articles/{id}/versions
func (as *ArticleService) GetArticleVersions(ctx context.Context, articleID int64) ([]params.ArticleVersionResponse, error) {

	userCanReadDraftedAndArchivedArticle, ok := ctx.Value(constanta.LocalUserCanReadDraftedAndArchivedArticle).(bool)
	if !ok {
		return nil, errors.New("error when parsing user permission")
	}

	statuses := []constanta.ArticleVersionStatus{constanta.Published}

	if userCanReadDraftedAndArchivedArticle {
		statuses = append(statuses, constanta.Draft, constanta.Archived)
	}

	articleVersions, err := as.ArticleRepo.GetArticleVersionsWithArticleIDAndStatuses(ctx, articleID, statuses...)
	if err != nil {
		return nil, err
	}

	res := make([]params.ArticleVersionResponse, len(articleVersions))
	for i, articleVersion := range articleVersions {
		res[i] = params.ArticleVersionResponse{
			VersionID: articleVersion.ID,
			Title:     articleVersion.Title,
			Body:      articleVersion.Body,
			Version:   articleVersion.Version,
			Status:    int8(articleVersion.Status),
			CreatedBy: articleVersion.CreatedBy,
			CreatedAt: articleVersion.CreatedAt,
			UpdatedBy: articleVersion.UpdatedBy,
			UpdatedAt: articleVersion.UpdatedAt,
		}
	}

	return res, nil
}

// TODO Pengambilan Daftar Artikel
// => GET /articles
