package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/elangreza/content-management-system/internal/entity"
)

type (
	ArticleRepo struct {
		db *sql.DB
	}
)

func NewArticleRepo(db *sql.DB) *ArticleRepo {
	return &ArticleRepo{
		db: db,
	}
}

const (
	createArticleQuery        = `INSERT INTO articles(created_by) VALUES($1) RETURNING id;`
	createArticleVersionQuery = `INSERT INTO article_versions
		(article_id, title, body, "version", status, created_by)
		VALUES($1, $2, $3, $4, $5, $6) RETURNING id;`
	updateLatestArticleVersionQuery = `UPDATE articles
		SET updated_by=$1, drafted_version_id=$2 WHERE id=$3;`
)

// CreateArticle implements ArticleRepo.
func (ar *ArticleRepo) CreateArticle(ctx context.Context, article entity.Article) (int, error) {
	var articleID int
	err := runInTx(ctx, ar.db, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx, createArticleQuery, article.CreatedBy).Scan(&articleID); err != nil {
			return err
		}

		versionID := 0
		for _, articleVersion := range article.Versions {
			if err := tx.QueryRowContext(ctx, createArticleVersionQuery,
				articleID,
				articleVersion.Title,
				articleVersion.Body,
				articleVersion.Version,
				articleVersion.Status,
				articleVersion.CreatedBy,
			).Scan(&versionID); err != nil {
				fmt.Println("2")
				return err
			}
		}

		if _, err := tx.ExecContext(ctx, updateLatestArticleVersionQuery,
			article.CreatedBy,
			versionID,
			articleID,
		); err != nil {
			fmt.Println("3")
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return articleID, nil
}
