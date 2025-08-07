package service

import (
	"context"
	"errors"

	"github.com/elangreza/content-management-system/internal/constanta"
	"github.com/elangreza/content-management-system/internal/entity"
	"github.com/elangreza/content-management-system/internal/params"
	"github.com/google/uuid"
)

type (
	articleRepo interface {
		CreateArticle(ctx context.Context, article entity.Article) (int, error)
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
