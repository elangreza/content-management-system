package constanta

type UserPermission int64

const (
	ReadDraftedAndArchivedArticle UserPermission = 1 << iota
	CreateArticle
	DeleteArticle
	UpdateStatusArticle
)
