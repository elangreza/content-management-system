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
	"github.com/google/uuid"
)

type (
	ArticleService interface {
		CreateArticle(ctx context.Context, req params.CreateArticleRequest) (*params.CreateArticleResponse, error)
		DeleteArticle(ctx context.Context, articleID int64) error
		UpdateStatusArticle(ctx context.Context, articleID, articleVersionID int64, status constanta.ArticleVersionStatus) error
		CreateArticleVersionWithReferenceFromArticleID(ctx context.Context, articleID int64, req params.CreateArticleVersionRequest) (*params.CreateArticleVersionResponse, error)
		CreateArticleVersionWithReferenceFromArticleIDAindVersionID(ctx context.Context, articleID int64, articleVersionID int64, req params.CreateArticleVersionRequest) (*params.CreateArticleVersionResponse, error)
		GetArticleWithID(ctx context.Context, articleID int64) (*params.GetArticleDetailResponse, error)
		GetArticleVersionWithIDAndArticleID(ctx context.Context, articleID int64, articleVersionID int64) (*params.ArticleVersionResponse, error)
		GetArticleVersions(ctx context.Context, articleID int64) ([]params.ArticleVersionResponse, error)
		GetArticles(ctx context.Context, req params.GetArticlesQueryParams) ([]params.ArticleVersionResponse, error)
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

func (ah *ArticleHandler) CreateNewArticleVersionWithReferenceFromArticleID(w http.ResponseWriter, r *http.Request) {
	articleIDParam := chi.URLParam(r, "articleID")

	articleID, err := strconv.Atoi(articleIDParam)
	if err != nil {
		err = errors.New("error when parsing articleID")
		sendErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	var body params.CreateArticleVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, errs.ValidationError{Message: err.Error()})
		return
	}

	newArticleVersion, err := ah.svc.CreateArticleVersionWithReferenceFromArticleID(r.Context(), int64(articleID), body)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	sendSuccessResponse(w, http.StatusOK, newArticleVersion)
}

func (ah *ArticleHandler) CreateNewArticleVersionWithReferenceFromArticleIDAndVersionID(w http.ResponseWriter, r *http.Request) {
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

	newArticleVersion, err := ah.svc.CreateArticleVersionWithReferenceFromArticleIDAindVersionID(r.Context(), int64(articleID), int64(articleVersionID), body)
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

func (ah *ArticleHandler) GetArticleVersionWithIDAndArticleID(w http.ResponseWriter, r *http.Request) {
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

	articleVersion, err := ah.svc.GetArticleVersionWithIDAndArticleID(r.Context(), int64(articleID), int64(articleVersionID))
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	sendSuccessResponse(w, http.StatusOK, articleVersion)
}

func (ah *ArticleHandler) GetArticleVersionsHandler(w http.ResponseWriter, r *http.Request) {
	articleIDParam := chi.URLParam(r, "articleID")
	articleID, err := strconv.Atoi(articleIDParam)
	if err != nil {
		err = errors.New("error when parsing articleID")
		sendErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	articleVersions, err := ah.svc.GetArticleVersions(r.Context(), int64(articleID))
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	sendSuccessResponse(w, http.StatusOK, articleVersions)
}

func (ah *ArticleHandler) GetArticlesHandler(w http.ResponseWriter, r *http.Request) {

	searchQuery := r.URL.Query().Get("search")
	sortQueries := r.URL.Query()["sorts"]
	limitQuery := r.URL.Query().Get("limit")
	pageQuery := r.URL.Query().Get("page")

	limit, _ := strconv.Atoi(limitQuery)
	page, _ := strconv.Atoi(pageQuery)
	queryParams := &params.GetArticlesQueryParams{
		Search: searchQuery,
		PaginationParams: params.PaginationParams{
			Sorts: sortQueries,
			Limit: limit,
			Page:  page,
		},
	}

	statusQueries := r.URL.Query()["status"]
	createdByQueries := r.URL.Query()["created_by"]
	updatedByQueries := r.URL.Query()["updated_by"]
	for _, v := range statusQueries {
		status, err := strconv.Atoi(v)
		if err != nil {
			sendErrorResponse(w, http.StatusBadRequest, errs.ValidationError{Message: "status must be an integer"})
			return
		}
		queryParams.Status = append(queryParams.Status, constanta.ArticleVersionStatus(status))
	}
	for _, v := range createdByQueries {
		createdBy, err := uuid.Parse(v)
		if err != nil {
			sendErrorResponse(w, http.StatusBadRequest, errs.ValidationError{Message: "not valid created_by uuid"})
			return
		}
		queryParams.CreatedBy = append(queryParams.CreatedBy, createdBy)
	}
	for _, v := range updatedByQueries {
		updatedBy, err := uuid.Parse(v)
		if err != nil {
			sendErrorResponse(w, http.StatusBadRequest, errs.ValidationError{Message: "not valid updated_by uuid"})
			return
		}
		queryParams.UpdatedBy = append(queryParams.UpdatedBy, updatedBy)
	}

	if err := queryParams.Validate(); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	articles, err := ah.svc.GetArticles(r.Context(), *queryParams)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	sendSuccessResponse(w, http.StatusOK, articles)
}
