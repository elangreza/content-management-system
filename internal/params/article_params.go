package params

type CreateArticleRequest struct {
	Title, Body string
}

type CreateArticleResponse struct {
	ArticleID int64
}

type UpdateArticleStatusRequest struct {
	Status int8 `json:"status"`
}

type CreateArticleVersionRequest struct {
	Title, Body string
}

type CreateArticleVersionResponse struct {
	ArticleVersionID int64
}
