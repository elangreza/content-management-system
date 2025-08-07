package entity

type UserPermission int

const (
	ReadDraftedArticle UserPermission = 1 << iota
	ReadArchivedArticle
	CreateArticle
	DeleteArticle
	UpdateStatusArticle
)

type UserRole struct {
	val         int
	permissions []UserPermission
}

var (
	ContentWriter UserRole = UserRole{
		permissions: []UserPermission{
			ReadDraftedArticle,
			ReadArchivedArticle,
			CreateArticle,
		},
	}

	Editor UserRole = UserRole{
		permissions: []UserPermission{
			ReadDraftedArticle,
			ReadArchivedArticle,
			CreateArticle,
			DeleteArticle,
			UpdateStatusArticle,
		},
	}
)

func (ur UserRole) GetPermissionValue() int {
	if len(ur.permissions) == 0 {
		return 0
	}

	var totalPermission UserPermission
	for _, permission := range ur.permissions {
		totalPermission += permission
	}

	return int(totalPermission)
}
