package sharevar

import (
	"github.com/elangreza/content-management-system/internal/constanta"
	"github.com/elangreza/content-management-system/internal/entity"
)

var (
	ContentWriter entity.UserRole = entity.NewUserRole(
		"ContentWriter",
		constanta.ReadDraftedArticle,
		constanta.ReadArchivedArticle,
		constanta.CreateArticle)

	Editor entity.UserRole = entity.NewUserRole(
		"Editor",
		constanta.ReadDraftedArticle,
		constanta.ReadArchivedArticle,
		constanta.CreateArticle,
		constanta.DeleteArticle,
		constanta.UpdateStatusArticle)
)
