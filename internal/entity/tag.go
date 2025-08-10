package entity

import "time"

type (
	Tag struct {
		Name string
	}

	ArticleVersionTag struct {
		TagName          string
		ArticleVersionID int64
	}

	TagUsage struct {
		Count         int
		LastUsed      time.Time
		TrendingScore float64
	}

	CalculateArticleVersionTagRelationShipScorePayload struct {
		Tags             []Tag
		ArticleVersionID int64
	}
)

func NewTags(tags ...string) []Tag {
	result := make([]Tag, len(tags))
	for i, tag := range tags {
		result[i] = Tag{Name: tag}
	}
	return result
}
