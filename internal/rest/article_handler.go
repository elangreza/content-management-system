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

// CreateArticleHandler
//
//	@Summary		Create a new article
//	@Description	Create a new article with the given parameters
//	@Tags			articles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string						true	"MUST HAVE PERMISSION CreateArticle. Fill with Bearer token. The token can be accessed via api /auth/login."
//	@Param			body			body		params.CreateArticleRequest	true	"Create Article Request"
//	@Success		201				{object}	params.CreateArticleResponse
//	@Failure		400				{object}	errs.ValidationError
//	@Failure		500				{object}	string
//	@Router			/articles [post]
func (ah *ArticleHandler) CreateArticleHandler(w http.ResponseWriter, r *http.Request) {

	body := params.CreateArticleRequest{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, errs.ValidationError{Message: err.Error()})
		return
	}

	if err := body.Validate(); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	Article, err := ah.svc.CreateArticle(r.Context(), body)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	sendSuccessResponse(w, http.StatusCreated, Article)
}

// DeleteArticleHandler
//
//	@Summary		Delete an article by ID
//	@Description	Delete an article with the given ID
//	@Tags			articles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string	true	"MUST HAVE PERMISSION DeleteArticle. Fill with Bearer token. The token can be accessed via api /auth/login."
//	@Success		200				{string}	string	"ok"
//	@Failure		400				{object}	errs.ValidationError
//	@Failure		500				{object}	object
//	@Router			/articles/{articleID} [delete]
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

// UpdateArticleStatusHandler
//
//	@Summary		Update the status of an article version
//	@Description	Update the status of an article version with the given parameters
//	@Tags			articles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization		header		string								true	"MUST HAVE PERMISSION UpdateStatusArticle. Fill with Bearer token. The token can be accessed via api /auth/login."
//	@Param			articleVersionID	path		int									true	"Article Version ID"
//	@Param			body				body		params.UpdateArticleStatusRequest	true	"Update Article Status Request"
//	@Success		200					{string}	string								"ok"
//	@Failure		400					{object}	errs.ValidationError
//	@Failure		500					{object}	string
//	@Failure		500					{object}	string
//	@Router			/articles/{articleID}/versions/{articleVersionID}/status [put]
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

// CreateNewArticleVersionWithReferenceFromArticleID
//
//	@Summary		Create a new article version with reference from an article ID
//	@Description	Create a new article version with reference from an article ID with the given parameters
//	@Tags			articles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string								true	"MUST HAVE PERMISSION CreateArticle. Fill with Bearer token. The token can be accessed via api /auth/login."
//	@Param			body			body		params.CreateArticleVersionRequest	true	"Create Article Version Request"
//	@Success		201				{object}	params.CreateArticleVersionResponse
//	@Failure		400				{object}	errs.ValidationError
//	@Failure		500				{object}	string
//	@Router			/articles/{articleID}/versions [post]
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

	sendSuccessResponse(w, http.StatusCreated, newArticleVersion)
}

// CreateNewArticleVersionWithReferenceFromArticleIDAndVersionID
//
//	@Summary		Create a new article version with reference from an article ID and version ID
//	@Description	Create a new article version with reference from an article ID and version ID with the given parameters
//	@Tags			articles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization		header		string								true	"MUST HAVE PERMISSION CreateArticle. Fill with Bearer token. The token can be accessed via api /auth/login."
//	@Param			articleVersionID	path		int									true	"Article Version ID"
//	@Param			body				body		params.CreateArticleVersionRequest	true	"Create Article Version Request"
//	@Success		201					{object}	params.CreateArticleVersionResponse
//	@Failure		400					{object}	errs.ValidationError
//	@Failure		500					{object}	string
//	@Router			/articles/{articleID}/versions/{articleVersionID} [post]
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

	sendSuccessResponse(w, http.StatusCreated, newArticleVersion)
}

// GetArticleDetailHandler
//
//	@Summary		Get article detail by ID
//	@Description	Get article detail with the given ID
//	@Tags			articles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string	false	"fill with Bearer token. The token can be accessed via api /auth/login. If authorization is not provided, the default behavior is showing only published articles. Otherwise, if the token is present and the user has permission to read drafted and archived articles, the token can be used to access draft, published, and archived articles. "//	@Param	articleID	path	int	true	"Article ID"
//	@Success		200				{object}	params.GetArticleDetailResponse
//	@Failure		400				{object}	errs.ValidationError
//	@Failure		500				{object}	object
//	@Router			/articles/{articleID} [get]
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

// GetArticleVersionWithIDAndArticleID
//
//	@Summary		Get article version by ID and article ID
//	@Description	Get article version with the given article ID and article version ID
//	@Tags			articles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization		header		string	false	"fill with Bearer token. The token can be accessed via api /auth/login. If authorization is not provided, the default behavior is showing only published articles. Otherwise, if the token is present and the user has permission to read drafted and archived articles, the token can be used to access draft, published, and archived articles. "//	@Param	articleID	path	int	true	"Article ID"
//	@Param			articleVersionID	path		int		true	"Article Version ID"
//	@Success		200					{object}	params.ArticleVersionResponse
//	@Failure		400					{object}	errs.ValidationError
//	@Failure		500					{object}	string
//	@Router			/articles/{articleID}/versions/{articleVersionID} [get]
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

// GetArticleVersionsHandler
//
//	@Summary		Get all versions of an article by ID
//	@Description	Get all versions of an article with the given ID
//	@Tags			articles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string	false	"fill with Bearer token. The token can be accessed via api /auth/login. If authorization is not provided, the default behavior is showing only published articles. Otherwise, if the token is present and the user has permission to read drafted and archived articles, the token can be used to access draft, published, and archived articles. "
//	@Success		200				{array}		params.ArticleVersionResponse
//	@Failure		400				{object}	errs.ValidationError
//	@Failure		500				{object}	object
//	@Router			/articles/{articleID}/versions [get]
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

// GetArticlesHandler
//
//	@Summary		Get all articles with optional filters
//	@Description	Get all articles with optional filters
//	@Tags			articles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string		false	"fill with Bearer token. The token can be accessed via api /auth/login. If authorization is not provided, the default behavior is showing only published articles. Otherwise, if the token is present and the user has permission to read drafted and archived articles, the token can be used to access draft, published, and archived articles. "
//	@Param			search			query		string		false	"Search query"
//	@Param			sorts			query		[]string	false	"article_id:asc | article_id:desc |	article_version_id:asc | article_version_id:desc |	created_by:asc | created_by:desc |	updated_by:asc | updated_by:desc |	title:asc | title:desc |	status:asc | status:desc |	version:asc | version:desc | created_at:asc | created_at:desc | updated_at:asc | updated_at:desc | tag_relationship_score:asc | tag_relationship_score:desc"
//	@Param			limit			query		int			false	"Limit"
//	@Param			page			query		int			false	"Page number"
//	@Param			status			query		int			false	"Status 0 for draft, 1 for published, 2 for archived (comma-separated, integer values)"
//	@Param			created_by		query		string		false	"Created by (comma-separated, UUID values)"
//	@Param			updated_by		query		string		false	"Updated by (comma-separated, UUID values)"
//	@Success		200				{array}		params.ArticleVersionResponse
//	@Failure		400				{object}	errs.ValidationError
//	@Failure		500				{object}	object
//	@Router			/articles [get]
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
