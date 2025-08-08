package postgresql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/elangreza/content-management-system/internal/constanta"
	"github.com/elangreza/content-management-system/internal/entity"
	"github.com/google/uuid"
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
				return err
			}
		}

		if _, err := tx.ExecContext(ctx, updateLatestArticleVersionQuery,
			article.CreatedBy,
			versionID,
			articleID,
		); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return articleID, nil
}

const (
	resetArticlePublishedAndDraftedToNullQuery = `UPDATE articles
		SET published_version_id=NULL, drafted_version_id=NULL WHERE id=$1;`
	deleteArticleVersionsByArticleIdQuery = `DELETE FROM article_versions
		WHERE article_id=$1;`
	deleteArticleByArticleIdQuery = `DELETE FROM articles
		WHERE id=$1;`
)

// DeleteArticle implements ArticleRepo.
func (ar *ArticleRepo) DeleteArticle(ctx context.Context, articleID int64) error {
	err := runInTx(ctx, ar.db, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, resetArticlePublishedAndDraftedToNullQuery, articleID); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, deleteArticleVersionsByArticleIdQuery, articleID); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, deleteArticleByArticleIdQuery, articleID); err != nil {
			return err
		}

		return nil
	})

	return err
}

const (
	getArticleVersionWithIDAndArticleIDQuery = `SELECT id, title, body, "version", status, created_by, created_at, updated_by, updated_at
		FROM article_versions WHERE article_id=$1 AND id=$2;`
)

func (ar *ArticleRepo) GetArticleVersionWithIDAndArticleID(ctx context.Context, articleID int64, articleVersionID int64) (*entity.ArticleVersion, error) {

	articleVersion := &entity.ArticleVersion{}
	updatedAt := sql.NullTime{}
	err := ar.db.QueryRowContext(ctx, getArticleVersionWithIDAndArticleIDQuery, articleID, articleVersionID).Scan(
		&articleVersion.ID,
		&articleVersion.Title,
		&articleVersion.Body,
		&articleVersion.Version,
		&articleVersion.Status,
		&articleVersion.CreatedBy,
		&articleVersion.CreatedAt,
		&articleVersion.UpdatedBy,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	if updatedAt.Valid {
		articleVersion.UpdatedAt = &updatedAt.Time
	}

	return articleVersion, nil
}

const (
	updateArticleVersionWithStatusPublishedIntoArchivedQuery = `UPDATE article_versions
		SET status=$1, updated_by=$2 WHERE article_id=$3;`
	updateArticleVersionQuery = `UPDATE article_versions
		SET status=$1, updated_by=$2 WHERE article_id=$3 AND id=$4;`
	updateArticlePublishedIdQuery = `UPDATE articles
		SET published_version_id=$1, updated_by=$2 WHERE id=$3;`
)

func (ar *ArticleRepo) UpdateArticleVersion(ctx context.Context, articleID, articleVersionID int64, status constanta.ArticleVersionStatus) error {

	localUserID, ok := ctx.Value(constanta.LocalUserID).(string)
	if !ok {
		return errors.New("error when handle ctx value")
	}

	userID, err := uuid.Parse(localUserID)
	if err != nil {
		return errors.New("error when parsing userID")
	}

	err = runInTx(ctx, ar.db, func(tx *sql.Tx) error {
		// update article version published into archived with article id
		if status == constanta.Published {
			if _, err := tx.ExecContext(ctx, updateArticleVersionWithStatusPublishedIntoArchivedQuery,
				constanta.Archived,
				userID,
				articleID,
			); err != nil {
				return err
			}
		}

		if _, err := tx.ExecContext(ctx, updateArticleVersionQuery,
			status,
			userID,
			articleID,
			articleVersionID,
		); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, updateArticlePublishedIdQuery,
			articleVersionID,
			userID,
			articleID,
		); err != nil {
			return err
		}

		return nil
	})

	return err
}
