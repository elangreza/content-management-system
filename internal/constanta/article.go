package constanta

type ArticleVersionStatus int8

const (
	Draft ArticleVersionStatus = iota
	Published
	Archived
)
