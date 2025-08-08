package constanta

type ArticleVersionStatus int8

const (
	Pending ArticleVersionStatus = iota
	Published
	Archived
)
