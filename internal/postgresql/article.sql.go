package postgresql

import (
	"context"
	"database/sql"

	"github.com/elangreza/content-management-system/internal/constanta"
	"github.com/elangreza/content-management-system/internal/entity"
	"github.com/google/uuid"
	"github.com/lib/pq"
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
		SET updated_by=$1, drafted_version_id=$2, version_sequence=$3 WHERE id=$4;`
)

// CreateArticle implements ArticleRepo.
func (ar *ArticleRepo) CreateArticle(ctx context.Context, article entity.Article) (int64, error) {
	var articleID int64
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
			article.VersionSequence,
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
		SET status=$1, updated_by=$2 WHERE article_id=$3 AND status=$4 RETURNING id;`
	updateArticleVersionQuery = `UPDATE article_versions
		SET status=$1, updated_by=$2 WHERE article_id=$3 AND id=$4;`
	updateArticlePublishedIdQuery = `UPDATE articles
		SET published_version_id=$1, updated_by=$2 WHERE id=$3;`
	updateArticleArchivedIdQuery = `UPDATE articles
		SET archived_version_id=$1, updated_by=$2 WHERE id=$3;`
)

func (ar *ArticleRepo) UpdateArticleStatus(ctx context.Context, articleID, articleVersionID int64, status constanta.ArticleVersionStatus, updatedBy uuid.UUID) error {
	err := runInTx(ctx, ar.db, func(tx *sql.Tx) error {
		// update article version published into archived with article id
		if status == constanta.Published {
			var archivedVersionID int64
			if err := tx.QueryRowContext(ctx, updateArticleVersionWithStatusPublishedIntoArchivedQuery,
				constanta.Archived,
				updatedBy,
				articleID,
				constanta.Published,
			).Scan(&archivedVersionID); err != nil {
				return err
			}

			if _, err := tx.ExecContext(ctx, updateArticleArchivedIdQuery,
				archivedVersionID,
				updatedBy,
				articleID,
			); err != nil {
				return err
			}
		}

		if _, err := tx.ExecContext(ctx, updateArticleVersionQuery,
			status,
			updatedBy,
			articleID,
			articleVersionID,
		); err != nil {
			return err
		}

		if status == constanta.Archived {
			if _, err := tx.ExecContext(ctx, updateArticleArchivedIdQuery,
				articleVersionID,
				updatedBy,
				articleID,
			); err != nil {
				return err
			}
		}

		if status == constanta.Published {
			if _, err := tx.ExecContext(ctx, updateArticlePublishedIdQuery,
				articleVersionID,
				updatedBy,
				articleID,
			); err != nil {
				return err
			}

			// Ensure the version exists
			// if there's no draft version, set the drafted_version_id to NULL
			// This is to ensure that the article has a valid draft version
			rows, err := tx.QueryContext(ctx, "SELECT id FROM article_versions WHERE status=$1 AND article_id = $2 ORDER BY version DESC LIMIT 1", constanta.Draft, articleID)
			if err != nil {
				return err
			}
			defer rows.Close()
			draftedVersions := make([]int64, 0)
			for rows.Next() {
				var draftVersionID int64
				if err := rows.Scan(&draftVersionID); err != nil {
					return err
				}
				draftedVersions = append(draftedVersions, draftVersionID)
			}
			if err := rows.Err(); err != nil {
				return err
			}

			// If there are pending versions, set the drafted_version_id to the first one
			if len(draftedVersions) > 0 {
				if _, err := tx.ExecContext(ctx, "UPDATE articles SET drafted_version_id=$1 WHERE id=$2", draftedVersions[0], articleID); err != nil {
					return err
				}
			} else {
				if _, err := tx.ExecContext(ctx, "UPDATE articles SET drafted_version_id=NULL WHERE id=$1", articleID); err != nil {
					return err
				}
			}
		}

		return nil
	})

	return err
}

func (ar *ArticleRepo) CreateArticleVersion(ctx context.Context, articleVersion entity.ArticleVersion) (int64, error) {
	var articleVersionID int64
	err := runInTx(ctx, ar.db, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx, createArticleVersionQuery,
			articleVersion.ArticleID,
			articleVersion.Title,
			articleVersion.Body,
			articleVersion.Version,
			articleVersion.Status,
			articleVersion.CreatedBy,
		).Scan(&articleVersionID); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, updateLatestArticleVersionQuery,
			articleVersion.CreatedBy,
			articleVersionID,
			articleVersion.Version,
			articleVersion.ArticleID,
		); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return articleVersionID, nil
}

const (
	getArticleWithIDQuery = `SELECT 
		id, 
		published_version_id, 
		drafted_version_id, 
		version_sequence, 
		created_by, 
		created_at, 
		updated_by, 
		updated_at
	FROM articles WHERE id=$1`
)

func (ar *ArticleRepo) GetArticleWithID(ctx context.Context, articleID int64) (*entity.Article, error) {
	article := &entity.Article{}
	publishedVersionID := sql.NullInt64{}
	draftedVersionID := sql.NullInt64{}
	err := ar.db.QueryRowContext(ctx, getArticleWithIDQuery, articleID).Scan(
		&article.ID,
		&publishedVersionID,
		&draftedVersionID,
		&article.VersionSequence,
		&article.CreatedBy,
		&article.CreatedAt,
		&article.UpdatedBy,
		&article.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if publishedVersionID.Valid {
		article.PublishedVersionID = publishedVersionID.Int64
	}
	if draftedVersionID.Valid {
		article.DraftedVersionID = draftedVersionID.Int64
	}

	return article, nil
}

const (
	getArticleVersionWithIDQuery = `SELECT 
		id, 
		article_id, 
		title, 
		body, 
		"version", 
		status, 
		created_by, 
		created_at,
		updated_by,
		updated_at
	FROM article_versions
	WHERE id=$1;`
)

func (ar *ArticleRepo) GetArticleVersionWithID(ctx context.Context, ID int64) (*entity.ArticleVersion, error) {
	articleVersion := &entity.ArticleVersion{}
	updatedAt := sql.NullTime{}
	err := ar.db.QueryRowContext(ctx, getArticleWithIDQuery, ID).Scan(
		&articleVersion.ID,
		&articleVersion.ArticleID,
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
	getArticleVersionsWithArticleIDAndStatusesQuery = `SELECT id, article_id, title, body, "version", status, created_by, created_at, updated_by, updated_at
	FROM article_versions WHERE article_id=$1 AND status = ANY($2) ORDER BY "version" DESC;`
)

func (ar *ArticleRepo) GetArticleVersionsWithArticleIDAndStatuses(ctx context.Context, articleID int64, status ...constanta.ArticleVersionStatus) ([]entity.ArticleVersion, error) {

	if len(status) == 0 {
		status = append(status, constanta.Published)
	}

	rows, err := ar.db.QueryContext(ctx, getArticleVersionsWithArticleIDAndStatusesQuery, articleID, pq.Array(status))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []entity.ArticleVersion
	for rows.Next() {
		var version entity.ArticleVersion
		updatedAt := sql.NullTime{}
		if err := rows.Scan(
			&version.ID,
			&version.ArticleID,
			&version.Title,
			&version.Body,
			&version.Version,
			&version.Status,
			&version.CreatedBy,
			&version.CreatedAt,
			&version.UpdatedBy,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		if updatedAt.Valid {
			version.UpdatedAt = &updatedAt.Time
		}
		versions = append(versions, version)
	}

	return versions, rows.Err()
}

func (ar *ArticleRepo) GetArticles(ctx context.Context, req entity.GetArticlesQueryServiceParams) ([]entity.ArticleVersion, error) {
	query := `SELECT 
			id, 
			article_id, 
			title, 
			body, 
			"version", 
			status, 
			created_by, 
			created_at, 
			updated_by, 
			updated_at
		FROM 
			article_versions
		WHERE 
			(status = ANY($1) OR $1 IS NULL)
		AND 
			(created_by = ANY($2) OR $2 IS NULL)
		AND 
			(updated_by = ANY($3) OR $3 IS NULL)
		AND 
			(title ILIKE '%' || $4 || '%' OR $4 IS NULL)
		ORDER BY ` + req.OrderClause + ` LIMIT $5 OFFSET $6;`

	offset := req.Limit * (req.Page - 1)

	rows, err := ar.db.QueryContext(ctx,
		query,
		pq.Array(req.Status),
		pq.Array(req.CreatedBy),
		pq.Array(req.UpdatedBy),
		req.Search,
		req.Limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articleVersions []entity.ArticleVersion
	for rows.Next() {
		var articleVersion entity.ArticleVersion
		updatedAt := sql.NullTime{}
		if err := rows.Scan(
			&articleVersion.ID,
			&articleVersion.ArticleID,
			&articleVersion.Title,
			&articleVersion.Body,
			&articleVersion.Version,
			&articleVersion.Status,
			&articleVersion.CreatedBy,
			&articleVersion.CreatedAt,
			&articleVersion.UpdatedBy,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		if updatedAt.Valid {
			articleVersion.UpdatedAt = &updatedAt.Time
		}
		articleVersions = append(articleVersions, articleVersion)
	}

	return articleVersions, rows.Err()
}
