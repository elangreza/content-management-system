package params

type CreateTagRequest struct {
	Names []string `json:"names"`
}

type GetTagResponse struct {
	Name          string  `json:"name"`
	UsageCount    int     `json:"usage_count"`
	TrendingScore float64 `json:"trending_score"`
}
