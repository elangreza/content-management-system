package rest

import (
	"context"
	"encoding/json"
	"net/http"

	errs "github.com/elangreza/content-management-system/internal/error"
	"github.com/elangreza/content-management-system/internal/params"
	"github.com/go-chi/chi/v5"
)

type (
	ArticleService interface {
		CreateArticle(ctx context.Context, req params.CreateArticleRequest) (*params.CreateArticleResponse, error)
	}

	ArticleHandler struct {
		svc ArticleService
	}
)

func NewArticleRouter(ar chi.Router, ArticleService ArticleService) {

	ArticleHandler := ArticleHandler{
		svc: ArticleService,
	}

	ar.Post("/article", ArticleHandler.ArticleUserHandler)
}

func (ah *ArticleHandler) ArticleUserHandler(w http.ResponseWriter, r *http.Request) {

	body := params.CreateArticleRequest{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, errs.ValidationError{Message: err.Error()})
		return
	}

	// TODO validation
	// if err := body.Validate(); err != nil {
	// 	Error(w, http.StatusBadRequest, err)
	// 	return
	// }

	Article, err := ah.svc.CreateArticle(r.Context(), body)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	sendSuccessResponse(w, http.StatusOK, Article)
}
