package constanta

type UserPermission int64

const (
	ReadDraftedArticle UserPermission = 1 << iota
	ReadArchivedArticle
	CreateArticle
	DeleteArticle
	UpdateStatusArticle
)
