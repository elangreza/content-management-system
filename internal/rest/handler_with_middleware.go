package rest

import (
	"github.com/elangreza/content-management-system/internal/constanta"
	"github.com/elangreza/content-management-system/internal/sharevar"
	"github.com/go-chi/chi/v5"
)

func NewHandlerWithMiddleware(
	publicRoute chi.Router,
	profileService ProfileService,
	authService AuthService,
	articleService ArticleService,
) {

	authMiddleware := AuthMiddleware{
		svc: authService,
	}

	articleMiddleware := ArticleMiddleware{
		svc: authService,
	}

	profileHandler := ProfileHandler{
		svc: profileService,
	}

	articleHandler := ArticleHandler{
		svc: articleService,
	}

	publicRoute.Group(func(r chi.Router) {
		r.Use(authMiddleware.MustAuthMiddleware())
		r.Get("/profile", profileHandler.ProfileUserHandler)

		r.Group(func(rCreateArticle chi.Router) {
			rCreateArticle.Use(authMiddleware.MustHavePermission(sharevar.ContentWriter.GetPermissions()...))
			rCreateArticle.Post("/articles", articleHandler.CreateArticleHandler)
			rCreateArticle.Post("/articles/{articleID}/versions/{articleVersionID}", articleHandler.CreateArticleVersionHandler)
		})

		r.Group(func(rDeletePermission chi.Router) {
			rDeletePermission.Use(authMiddleware.MustHavePermission(constanta.DeleteArticle))
			rDeletePermission.Delete("/articles/{articleID}", articleHandler.DeleteArticleHandler)
		})

		r.Group(func(rUpdateStatusPermission chi.Router) {
			rUpdateStatusPermission.Use(authMiddleware.MustHavePermission(constanta.UpdateStatusArticle))
			rUpdateStatusPermission.Put("/articles/{articleID}/versions/{articleVersionID}/status", articleHandler.UpdateArticleStatusHandler)
		})

	})

	publicRoute.Group(func(r chi.Router) {
		r.Use(authMiddleware.OptionalAuthMiddleware())
		r.Use(articleMiddleware.CanSeeDraftOrArchivedArticle())
		r.Get("/articles/{articleID}", articleHandler.GetArticleDetailHandler)
		r.Get("/articles/{articleID}/versions", articleHandler.GetArticleVersionsHandler)
		r.Get("/articles/{articleID}/versions/{articleVersionID}", articleHandler.GetArticleVersionWithIDAndArticleID)
	})
}
