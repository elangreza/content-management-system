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
func (as *ArticleService) UpdateStatusArticle(ctx context.Context, articleID, articleVersionID int64, status constanta.ArticleVersionStatus) error {
	userID, ok := ctx.Value(constanta.LocalUserID).(uuid.UUID)
	if !ok {
		return errors.New("error when parsing userID")
	}

	articleVersion, err := as.ArticleRepo.GetArticleVersionWithIDAndArticleID(ctx, articleID, articleVersionID)
	if err != nil {
		return err
	}

	if status == articleVersion.Status {
		return errs.ValidationError{Message: "status cannot be same as current status"}
	}

	if status < articleVersion.Status {
		return errs.ValidationError{Message: "status cannot be downgraded"}
	}

	return as.ArticleRepo.UpdateArticleVersion(ctx, articleID, articleVersionID, status, userID)
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

// TODO Pengambilan Daftar Artikel
// => GET /articles
// func (as *ArticleService) GetArticles(ctx context.Context, param params.GetParamRequest) (*entity.Article, error) {
// 	return nil, nil
// }

// TODO Pengambilan Detail Artikel Terbaru
// => POST /articles/{id}
// TODO Pengambilan Daftar Versi Artikel
// => GET /articles/{id}/versions
// TODO Pengambilan Detail Versi Artikel Tertentu
// => GET /articles/{id}/versions/{id}
