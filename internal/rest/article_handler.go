package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/elangreza/content-management-system/internal/constanta"
	errs "github.com/elangreza/content-management-system/internal/error"
	"github.com/elangreza/content-management-system/internal/params"
	"github.com/go-chi/chi/v5"
)

type (
	ArticleService interface {
		CreateArticle(ctx context.Context, req params.CreateArticleRequest) (*params.CreateArticleResponse, error)
		DeleteArticle(ctx context.Context, articleID int64) error
		UpdateStatusArticle(ctx context.Context, articleID, articleVersionID int64, status constanta.ArticleVersionStatus) error
		CreateArticleVersion(ctx context.Context, articleID int64, articleVersionID int64, req params.CreateArticleVersionRequest) (*params.CreateArticleVersionResponse, error)
		GetArticleWithID(ctx context.Context, articleID int64) (*params.GetArticleDetailResponse, error)
	}

	ArticleHandler struct {
		svc ArticleService
	}
)

func (ah *ArticleHandler) CreateArticleHandler(w http.ResponseWriter, r *http.Request) {

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

func (ah *ArticleHandler) DeleteArticleHandler(w http.ResponseWriter, r *http.Request) {

	articleIDParam := chi.URLParam(r, "articleID")

	articleID, err := strconv.Atoi(articleIDParam)
	if err != nil {
		err = errors.New("error when parsing articleID")
		sendErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	err = ah.svc.DeleteArticle(r.Context(), int64(articleID))
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	sendSuccessResponse(w, http.StatusOK, "ok")
}

func (ah *ArticleHandler) UpdateArticleStatusHandler(w http.ResponseWriter, r *http.Request) {
	articleIDParam := chi.URLParam(r, "articleID")

	articleID, err := strconv.Atoi(articleIDParam)
	if err != nil {
		err = errors.New("error when parsing articleID")
		sendErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	articleVersionIDParam := chi.URLParam(r, "articleVersionID")
	articleVersionID, err := strconv.Atoi(articleVersionIDParam)
	if err != nil {
		err = errors.New("error when parsing articleVersionID")
		sendErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	var body params.UpdateArticleStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, errs.ValidationError{Message: err.Error()})
		return
	}

	if body.Status < 0 || body.Status > 3 {
		sendErrorResponse(w, http.StatusBadRequest, errs.ValidationError{Message: "status must be between 0 and 3"})
		return
	}

	err = ah.svc.UpdateStatusArticle(r.Context(), int64(articleID), int64(articleVersionID), constanta.ArticleVersionStatus(body.Status))
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	sendSuccessResponse(w, http.StatusOK, "ok")
}

func (ah *ArticleHandler) CreateArticleVersionHandler(w http.ResponseWriter, r *http.Request) {
	articleIDParam := chi.URLParam(r, "articleID")
	articleVersionIDParam := chi.URLParam(r, "articleVersionID")

	articleID, err := strconv.Atoi(articleIDParam)
	if err != nil {
		err = errors.New("error when parsing articleID")
		sendErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	articleVersionID, err := strconv.Atoi(articleVersionIDParam)
	if err != nil {
		err = errors.New("error when parsing articleVersionID")
		sendErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	var body params.CreateArticleVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, errs.ValidationError{Message: err.Error()})
		return
	}

	newArticleVersion, err := ah.svc.CreateArticleVersion(r.Context(), int64(articleID), int64(articleVersionID), body)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	sendSuccessResponse(w, http.StatusOK, newArticleVersion)
}

func (ah *ArticleHandler) GetArticleDetailHandler(w http.ResponseWriter, r *http.Request) {
	articleIDParam := chi.URLParam(r, "articleID")

	articleID, err := strconv.Atoi(articleIDParam)
	if err != nil {
		err = errors.New("error when parsing articleID")
		sendErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	articleDetail, err := ah.svc.GetArticleWithID(r.Context(), int64(articleID))
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	sendSuccessResponse(w, http.StatusOK, articleDetail)
}
