package params

type CreateArticleRequest struct {
	Title, Body string
}

type CreateArticleResponse struct {
	ID int
}

type UpdateArticleStatusRequest struct {
	Status int8 `json:"status"` // e.g., 0 for draft, 1 for published, 2 for archived, 3 for deleted
}
