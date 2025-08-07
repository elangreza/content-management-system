package params

type CreateArticleRequest struct {
	Title, Body string
}

type CreateArticleResponse struct {
	ID int
}
