package params

import (
	"time"

	errs "github.com/elangreza/content-management-system/internal/error"
)

type CreateTagRequest struct {
	Names []string `json:"names"`
}

func (ctr *CreateTagRequest) Validate() error {
	if len(ctr.Names) == 0 {
		return errs.ValidationError{Message: "at least one tag name is required"}
	}

	for _, name := range ctr.Names {
		if name == "" {
			return errs.ValidationError{Message: "tag name cannot be empty"}
		}
	}

	return nil
}

type GetTagResponse struct {
	Name          string    `json:"name"`
	UsageCount    int       `json:"usage_count"`
	TrendingScore float64   `json:"trending_score"`
	LastUsed      time.Time `json:"last_used"`
}

type GetTagsRequest struct {
	SortValue string
	Direction string
}
