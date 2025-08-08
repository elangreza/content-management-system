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
		CreateArticle(ctx context.Context, article entity.Article) (int, error)
		DeleteArticle(ctx context.Context, articleID int64) error
		GetArticleVersionWithIDAndArticleID(ctx context.Context, articleID int64, articleVersionID int64) (*entity.ArticleVersion, error)
		UpdateArticleVersion(ctx context.Context, articleID int64, articleVersionID int64, status constanta.ArticleVersionStatus) error
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

	localUserID, ok := ctx.Value(constanta.LocalUserID).(string)
	if !ok {
		return nil, errors.New("error when handle ctx value")
	}

	userID, err := uuid.Parse(localUserID)
	if err != nil {
		return nil, errors.New("error when parsing userID")
	}

	article := entity.NewArticle(req.Title, req.Body, userID)
	id, err := as.ArticleRepo.CreateArticle(ctx, *article)
	if err != nil {
		return nil, err
	}

	return &params.CreateArticleResponse{
		ID: id,
	}, nil
}

// => DELETE /articles/{id}
func (as *ArticleService) DeleteArticle(ctx context.Context, articleID int64) error {
	return as.ArticleRepo.DeleteArticle(ctx, articleID)
}

// => PUT /articles/{id}/versions/{id}/status
func (as *ArticleService) UpdateStatusArticle(ctx context.Context, articleID, articleVersionID int64, status constanta.ArticleVersionStatus) error {
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

	return as.ArticleRepo.UpdateArticleVersion(ctx, articleID, articleVersionID, status)
}

// DONE Pembuatan Artikel Baru
// => POST /articles
// TODO Pengambilan Daftar Artikel
// => GET /articles
// TODO Pengambilan Detail Artikel Terbaru
// => POST /articles/{id}
// TODO Pembuatan Versi Artikel Baru
// => POST /articles/{id}/versions/{id}
// DONE Penghapusan Artikel
// => DELETE /articles/{id}
// DONE Perubahan Status Versi Artikel
// => PUT /articles/{id}/versions/{id}/status
// TODO Pengambilan Daftar Versi Artikel
// => GET /articles/{id}/versions
// TODO Pengambilan Detail Versi Artikel Tertentu
// => GET /articles/{id}/versions/{id}
