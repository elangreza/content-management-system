package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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
	createArticleVersionTagsQuery = `INSERT INTO article_version_tags (article_version_id, tag_name)
		VALUES ($1,$2) ON CONFLICT (article_version_id, tag_name) DO NOTHING;`
)

func (ar *ArticleRepo) CreateArticle(ctx context.Context, article entity.Article, articleVersion entity.ArticleVersion) (int64, int64, error) {
	var articleID, articleVersionID int64
	err := runInTx(ctx, ar.db, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx, createArticleQuery, article.CreatedBy).Scan(&articleID); err != nil {
			return err
		}

		if err := tx.QueryRowContext(ctx, createArticleVersionQuery,
			articleID,
			articleVersion.Title,
			articleVersion.Body,
			articleVersion.Version,
			articleVersion.Status,
			articleVersion.CreatedBy,
		).Scan(&articleVersionID); err != nil {
			return err
		}

		if len(articleVersion.Tags) > 0 {

			// insert the tag relationship for this article version
			for _, tag := range articleVersion.Tags {
				_, err := tx.ExecContext(ctx, upsertTagQuery, tag.Name)
				if err != nil {
					return err
				}

				_, err = tx.ExecContext(ctx, createArticleVersionTagsQuery, articleVersionID, tag.Name)
				if err != nil {
					return err
				}
			}
		}

		if _, err := tx.ExecContext(ctx, updateLatestArticleVersionQuery,
			article.CreatedBy,
			articleVersionID,
			article.VersionSequence,
			articleID,
		); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, 0, err
	}

	return articleID, articleVersionID, nil
}

const (
	resetArticlePublishedAndDraftedToNullQuery = `UPDATE articles
		SET published_version_id=NULL, drafted_version_id=NULL WHERE id=$1;`
	deleteArticleVersionsByArticleIdQuery = `DELETE FROM article_versions
		WHERE article_id=$1;`
	deleteArticleByArticleIdQuery = `DELETE FROM articles
		WHERE id=$1;`
)

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
	getArticleVersionWithIDAndArticleIDQuery = `SELECT
		id,
		article_id,
		title,
		body,
		"version",
		status,
		tag_relationship_score,
		created_by,
		created_at,
		updated_by,
		updated_at
	FROM 
		article_versions WHERE article_id=$1 AND id=$2;`
)

func (ar *ArticleRepo) GetArticleVersionWithIDAndArticleID(ctx context.Context, articleID int64, articleVersionID int64) (*entity.ArticleVersion, error) {

	articleVersion := &entity.ArticleVersion{}
	updatedAt := sql.NullTime{}
	err := ar.db.QueryRowContext(ctx, getArticleVersionWithIDAndArticleIDQuery, articleID, articleVersionID).Scan(
		&articleVersion.ArticleVersionID,
		&articleVersion.ArticleID,
		&articleVersion.Title,
		&articleVersion.Body,
		&articleVersion.Version,
		&articleVersion.Status,
		&articleVersion.TagRelationShipScore,
		&articleVersion.CreatedBy,
		&articleVersion.CreatedAt,
		&articleVersion.UpdatedBy,
		&articleVersion.UpdatedAt,
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
		SET status=$1, updated_by=$2 WHERE article_id=$3 AND status=$4;`
	updateArticleVersionQuery = `UPDATE article_versions
		SET status=$1, updated_by=$2 WHERE article_id=$3 AND id=$4;`
	updateArticlePublishedIdQuery = `UPDATE articles
		SET published_version_id=$1, updated_by=$2 WHERE id=$3;`
	updateArticleArchivedIdQuery = `UPDATE articles
		SET archived_version_id=$1, updated_by=$2 WHERE id=$3;`
)

func (ar *ArticleRepo) UpdateArticleStatus(ctx context.Context, articleID, articleVersionID int64, status, prevStatus constanta.ArticleVersionStatus, updatedBy uuid.UUID) error {
	err := runInTx(ctx, ar.db, func(tx *sql.Tx) error {
		// update article version published into archived with article id
		if status == constanta.Published {

			row, err := tx.QueryContext(ctx, "SELECT id FROM article_versions WHERE article_id = $1 AND status = $2 LIMIT 1", articleID, constanta.Published)
			if err != nil {
				return err
			}
			defer row.Close()
			var existingPublishedVersionID int64
			for row.Next() {
				err := row.Scan(&existingPublishedVersionID)
				if err != nil {
					return err
				}
			}

			if err := row.Err(); err != nil {
				return err
			}

			if existingPublishedVersionID != 0 {
				_, err = tx.ExecContext(ctx, updateArticleVersionWithStatusPublishedIntoArchivedQuery,
					constanta.Archived,
					updatedBy,
					articleID,
					constanta.Published,
				)
				if err != nil {
					return err
				}

				// update existing published version id to archived version id
				if _, err := tx.ExecContext(ctx, updateArticleArchivedIdQuery,
					existingPublishedVersionID,
					updatedBy,
					articleID,
				); err != nil {
					return err
				}
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

			if prevStatus == constanta.Published {

				// if the previous status is published, set the published_version_id to NULL
				if _, err := tx.ExecContext(ctx, updateArticlePublishedIdQuery, nil, updatedBy, articleID); err != nil {
					return err
				}
			}

			if prevStatus == constanta.Draft {
				// if the previous status is draft search latest draft version
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

const (
	deleteArticleVersionTags = `DELETE FROM article_version_tags WHERE article_version_id = $1`
)

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

		if len(articleVersion.Tags) > 0 {

			// delete existing tag relationships for this article version
			_, err := tx.ExecContext(ctx, `DELETE FROM article_version_tags WHERE article_version_id = $1`, articleVersionID)
			if err != nil {
				return err
			}

			// insert the tag relationship for this article version
			for _, tag := range articleVersion.Tags {
				_, err := tx.ExecContext(ctx, upsertTagQuery, tag.Name)
				if err != nil {
					return err
				}

				_, err = tx.ExecContext(ctx, createArticleVersionTagsQuery, articleVersionID, tag.Name)
				if err != nil {
					return err
				}
			}
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
		archived_version_id, 
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
	archivedVersionID := sql.NullInt64{}
	err := ar.db.QueryRowContext(ctx, getArticleWithIDQuery, articleID).Scan(
		&article.ID,
		&publishedVersionID,
		&draftedVersionID,
		&archivedVersionID,
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
	if archivedVersionID.Valid {
		article.ArchivedVersionID = archivedVersionID.Int64
	}

	return article, nil
}

const (
	getArticleVersionsWithArticleIDAndStatusesQuery = `SELECT 
	id, 
	article_id, 
	title, 
	body, 
	"version", 
	status,
	tag_relationship_score,
	created_by, 
	created_at, 
	updated_by, 
	updated_at
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
			&version.ArticleVersionID,
			&version.ArticleID,
			&version.Title,
			&version.Body,
			&version.Version,
			&version.Status,
			&version.TagRelationShipScore,
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

func getArticleQueryByStatus(status constanta.ArticleVersionStatus) string {
	column := "published_version_id"
	switch status {
	case constanta.Draft:
		column = "drafted_version_id"
	case constanta.Archived:
		column = "archived_version_id"
	}
	q := `SELECT
			av.id as article_version_id, 
			av.article_id as article_id, 
			av.title as title, 
			av.body as body, 
			av.version as "version", 
			av.status as status, 
			av.tag_relationship_score as tag_relationship_score,
			av.created_by as created_by, 
			av.created_at as created_at, 
			av.updated_by as updated_by, 
			av.updated_at as updated_at
		FROM public.articles a
		JOIN public.article_versions av ON a.%s = av.id WHERE a.%s IS NOT NULL AND av.status = %d`

	return fmt.Sprintf(q, column, column, int8(status))
}

func (ar *ArticleRepo) GetArticles(ctx context.Context, req entity.GetArticlesQueryServiceParams) ([]entity.ArticleVersion, error) {
	unionQuery := []string{}
	for _, v := range req.Status {
		q := getArticleQueryByStatus(v)
		unionQuery = append(unionQuery, q)
	}

	query := "select * from (" + strings.Join(unionQuery, " UNION ALL ") + ")" +
		` WHERE
			(created_by = ANY($1) OR $1 IS NULL)
		AND 
			(updated_by = ANY($2) OR $2 IS NULL)
		AND 
			(title ILIKE '%' || $3 || '%' OR $3 IS NULL)
		ORDER BY ` + req.OrderClause + ` LIMIT $4 OFFSET $5;`

	offset := req.Limit * (req.Page - 1)

	rows, err := ar.db.QueryContext(ctx,
		query,
		// pq.Array(req.Status),
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
			&articleVersion.ArticleVersionID,
			&articleVersion.ArticleID,
			&articleVersion.Title,
			&articleVersion.Body,
			&articleVersion.Version,
			&articleVersion.Status,
			&articleVersion.TagRelationShipScore,
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

const (
	getTagsWithArticleVersionIDQuery = `SELECT tag_name FROM article_version_tags WHERE article_version_id = $1 ORDER BY tag_name asc;`
)

func (ar *ArticleRepo) GetTagsWithArticleVersionID(ctx context.Context, articleVersionID int64) ([]entity.Tag, error) {
	rows, err := ar.db.QueryContext(ctx, getTagsWithArticleVersionIDQuery, articleVersionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []entity.Tag
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, entity.Tag{Name: tag})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

const (
	UpdateArticleVersionRelationshipScoreQuery = `UPDATE article_versions SET tag_relationship_score = $1 WHERE id = $2;`
)

func (ar *ArticleRepo) UpdateArticleVersionRelationshipScore(ctx context.Context, articleVersionID int64, relationshipScore float64) error {
	_, err := ar.db.ExecContext(ctx, UpdateArticleVersionRelationshipScoreQuery, relationshipScore, articleVersionID)
	if err != nil {
		return err
	}

	return nil
}
