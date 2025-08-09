package service

import (
	"context"
	"errors"
	"reflect"
	"slices"

	"github.com/elangreza/content-management-system/internal/constanta"
	"github.com/elangreza/content-management-system/internal/entity"
	errs "github.com/elangreza/content-management-system/internal/error"
	"github.com/elangreza/content-management-system/internal/params"
	"github.com/google/uuid"
)

type (
	articleRepo interface {
		CreateArticle(ctx context.Context, article entity.Article, articleVersion entity.ArticleVersion) (int64, error)
		DeleteArticle(ctx context.Context, articleID int64) error
		GetArticleVersionWithIDAndArticleID(ctx context.Context, articleID int64, articleVersionID int64) (*entity.ArticleVersion, error)
		UpdateArticleStatus(ctx context.Context, articleID int64, articleVersionID int64, status constanta.ArticleVersionStatus, updatedBy uuid.UUID) error
		CreateArticleVersion(ctx context.Context, articleVersion entity.ArticleVersion) (int64, error)
		GetArticleWithID(ctx context.Context, articleID int64) (*entity.Article, error)
		GetArticleVersionsWithArticleIDAndStatuses(ctx context.Context, ArticleID int64, status ...constanta.ArticleVersionStatus) ([]entity.ArticleVersion, error)
		GetArticles(ctx context.Context, req entity.GetArticlesQueryServiceParams) ([]entity.ArticleVersion, error)
		GetRawTagsWithArticleVersionID(ctx context.Context, articleVersionID int64) ([]string, error)
	}

	ArticleService struct {
		articleRepo articleRepo
	}
)

func NewArticleService(articleRepo articleRepo) *ArticleService {
	return &ArticleService{
		articleRepo: articleRepo,
	}
}

// => POST /articles
func (as *ArticleService) CreateArticle(ctx context.Context, req params.CreateArticleRequest) (*params.CreateArticleResponse, error) {
	userID, ok := ctx.Value(constanta.LocalUserID).(uuid.UUID)
	if !ok {
		return nil, errors.New("error when parsing userID")
	}

	article := entity.NewArticle(req.Title, req.Body, userID)
	articleVersion := entity.NewArticleVersion(article.ID, req.Title, req.Body, userID, 1, req.Tags)
	id, err := as.articleRepo.CreateArticle(ctx, *article, *articleVersion)
	if err != nil {
		return nil, err
	}

	return &params.CreateArticleResponse{
		ArticleID: id,
	}, nil
}

// => DELETE /articles/{id}
func (as *ArticleService) DeleteArticle(ctx context.Context, articleID int64) error {
	return as.articleRepo.DeleteArticle(ctx, articleID)
}

// => PUT /articles/{id}/versions/{id}/status
func (as *ArticleService) UpdateStatusArticle(ctx context.Context, articleID, articleVersionID int64, reqStatus constanta.ArticleVersionStatus) error {
	userID, ok := ctx.Value(constanta.LocalUserID).(uuid.UUID)
	if !ok {
		return errors.New("error when parsing userID")
	}

	articleVersion, err := as.articleRepo.GetArticleVersionWithIDAndArticleID(ctx, articleID, articleVersionID)
	if err != nil {
		return err
	}

	if reqStatus == articleVersion.Status {
		return errs.ValidationError{Message: "status cannot be same as current status"}
	}

	if reqStatus < articleVersion.Status {
		return errs.ValidationError{Message: "status cannot be downgraded"}
	}

	return as.articleRepo.UpdateArticleStatus(ctx, articleID, articleVersionID, reqStatus, userID)
}

// => PUT /articles/{articleID}
func (as *ArticleService) CreateArticleVersionWithReferenceFromArticleID(ctx context.Context, articleID int64, req params.CreateArticleVersionRequest) (*params.CreateArticleVersionResponse, error) {
	userID, ok := ctx.Value(constanta.LocalUserID).(uuid.UUID)
	if !ok {
		return nil, errors.New("error when parsing userID")
	}

	article, err := as.articleRepo.GetArticleWithID(ctx, articleID)
	if err != nil {
		return nil, err
	}

	// get the latest version ID if article has a drafted version
	var articleVersionID int64
	if article.DraftedVersionID != 0 {
		articleVersionID = article.DraftedVersionID
	}

	// if articleVersionID is 0, use the published version ID if it exists
	if articleVersionID == 0 && article.PublishedVersionID != 0 {
		articleVersionID = article.PublishedVersionID
	}

	// if articleVersionID is existing, check if the new version is the same as the current version
	if articleVersionID != 0 {
		articleVersion, err := as.articleRepo.GetArticleVersionWithIDAndArticleID(ctx, articleID, articleVersionID)
		if err != nil {
			return nil, err
		}

		tags, err := as.articleRepo.GetRawTagsWithArticleVersionID(ctx, articleVersionID)
		if err != nil {
			return nil, err
		}

		slices.Sort(req.Tags)

		if articleVersion.Title == req.Title && articleVersion.Body == req.Body && reflect.DeepEqual(tags, req.Tags) {
			return nil, errs.ValidationError{Message: "title, tags and body cannot be the same as the current version"}
		}
	}

	version := article.VersionSequence + 1
	newArticleVersion := entity.NewArticleVersion(articleID, req.Title, req.Body, userID, version, req.Tags)
	newArticleVersionID, err := as.articleRepo.CreateArticleVersion(ctx, *newArticleVersion)
	if err != nil {
		return nil, err
	}

	return &params.CreateArticleVersionResponse{
		ArticleVersionID: newArticleVersionID,
	}, nil
}

// => PUT /articles/{id}/versions/{id}
func (as *ArticleService) CreateArticleVersionWithReferenceFromArticleIDAindVersionID(ctx context.Context, articleID int64, articleVersionID int64, req params.CreateArticleVersionRequest) (*params.CreateArticleVersionResponse, error) {
	userID, ok := ctx.Value(constanta.LocalUserID).(uuid.UUID)
	if !ok {
		return nil, errors.New("error when parsing userID")
	}

	article, err := as.articleRepo.GetArticleWithID(ctx, articleID)
	if err != nil {
		return nil, err
	}

	articleVersion, err := as.articleRepo.GetArticleVersionWithIDAndArticleID(ctx, articleID, articleVersionID)
	if err != nil {
		return nil, err
	}

	tags, err := as.articleRepo.GetRawTagsWithArticleVersionID(ctx, articleVersionID)
	if err != nil {
		return nil, err
	}

	slices.Sort(req.Tags)

	if articleVersion.Title == req.Title && articleVersion.Body == req.Body && reflect.DeepEqual(tags, req.Tags) {
		return nil, errs.ValidationError{Message: "title, tags and body cannot be the same as the current version"}
	}

	version := article.VersionSequence + 1
	newArticleVersion := entity.NewArticleVersion(articleID, req.Title, req.Body, userID, version, req.Tags)
	newArticleVersionID, err := as.articleRepo.CreateArticleVersion(ctx, *newArticleVersion)
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
// Pengambilan Daftar Versi Artikel
// => GET /articles/{id}/versions

// => POST /articles/{id}
func (as *ArticleService) GetArticleWithID(ctx context.Context, articleID int64) (*params.GetArticleDetailResponse, error) {
	article, err := as.articleRepo.GetArticleWithID(ctx, articleID)
	if err != nil {
		return nil, err
	}

	userCanReadDraftedAndArchivedArticle, ok := ctx.Value(constanta.LocalUserCanReadDraftedAndArchivedArticle).(bool)
	if !ok {
		return nil, errors.New("error when parsing user permission")
	}

	var articleVersionResponse *params.ArticleVersionResponse
	if article.DraftedVersionID != 0 && userCanReadDraftedAndArchivedArticle {
		draftedVersion, err := as.articleRepo.GetArticleVersionWithIDAndArticleID(ctx, article.ID, article.DraftedVersionID)
		if err != nil {
			return nil, err
		}

		articleVersionResponse = &params.ArticleVersionResponse{
			ArticleID: draftedVersion.ArticleID,
			VersionID: draftedVersion.ArticleVersionID,
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
		publishedVersion, err := as.articleRepo.GetArticleVersionWithIDAndArticleID(ctx, article.ID, article.PublishedVersionID)
		if err != nil {
			return nil, err
		}

		publishedVersionResponse = &params.ArticleVersionResponse{
			ArticleID: publishedVersion.ArticleID,
			VersionID: publishedVersion.ArticleVersionID,
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
	articleVersion, err := as.articleRepo.GetArticleVersionWithIDAndArticleID(ctx, articleID, articleVersionID)
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
		ArticleID: articleVersion.ArticleID,
		VersionID: articleVersion.ArticleVersionID,
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

	articleVersions, err := as.articleRepo.GetArticleVersionsWithArticleIDAndStatuses(ctx, articleID, statuses...)
	if err != nil {
		return nil, err
	}

	res := make([]params.ArticleVersionResponse, len(articleVersions))
	for i, articleVersion := range articleVersions {
		res[i] = params.ArticleVersionResponse{
			ArticleID: articleVersion.ArticleID,
			VersionID: articleVersion.ArticleVersionID,
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
func (as *ArticleService) GetArticles(ctx context.Context, req params.GetArticlesQueryParams) ([]params.ArticleVersionResponse, error) {

	userCanReadDraftedAndArchivedArticle, ok := ctx.Value(constanta.LocalUserCanReadDraftedAndArchivedArticle).(bool)
	if !ok {
		return nil, errors.New("error when parsing user permission")
	}

	if !userCanReadDraftedAndArchivedArticle {
		req.Status = []constanta.ArticleVersionStatus{constanta.Published}
	}

	articleVersions, err := as.articleRepo.GetArticles(ctx, entity.GetArticlesQueryServiceParams{
		Search:      req.Search,
		Status:      req.Status,
		CreatedBy:   req.CreatedBy,
		UpdatedBy:   req.UpdatedBy,
		OrderClause: req.GetOrderClause(),
		Limit:       req.Limit,
		Page:        req.Page,
	})

	if err != nil {
		return nil, err
	}

	res := make([]params.ArticleVersionResponse, len(articleVersions))
	for i, articleVersion := range articleVersions {
		res[i] = params.ArticleVersionResponse{
			ArticleID: articleVersion.ArticleID,
			VersionID: articleVersion.ArticleVersionID,
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

func (as *ArticleService) GetRawTagsWithArticleVersionID(ctx context.Context, articleVersionID int64) ([]string, error) {
	return as.articleRepo.GetRawTagsWithArticleVersionID(ctx, articleVersionID)
}
