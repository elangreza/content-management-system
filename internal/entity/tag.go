package entity

type (
	Tag struct {
		Name string
	}

	ArticleVersionTag struct {
		TagName          string
		ArticleVersionID int64
	}
)
