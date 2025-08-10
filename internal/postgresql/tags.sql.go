package postgresql

import (
	"context"
	"database/sql"
	"time"

	"github.com/elangreza/content-management-system/internal/constanta"
	"github.com/elangreza/content-management-system/internal/entity"
)

type (
	TagsRepo struct {
		db *sql.DB
	}
)

func NewTagRepo(db *sql.DB) *TagsRepo {
	return &TagsRepo{
		db: db,
	}
}

const (
	upsertTagQuery = `INSERT INTO tags ("name") VALUES ($1) ON CONFLICT (name) DO NOTHING`
)

// UpsertTags implements TagsRepo.
func (u *TagsRepo) UpsertTags(ctx context.Context, names ...string) error {
	preparedQuery, err := u.db.Prepare(upsertTagQuery)
	if err != nil {
		if err != nil {
			return err
		}
	}

	for _, v := range names {
		_, err = preparedQuery.ExecContext(ctx, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *TagsRepo) GetTags(ctx context.Context) ([]string, error) {
	query := `SELECT name FROM tags`
	rows, err := u.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (u *TagsRepo) GetTagUsageCounts(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT t.name, COUNT(avt.tag_name) AS usage_count
		FROM tags t
		LEFT JOIN article_version_tags avt ON t.name = avt.tag_name 
		LEFT JOIN article_versions av ON avt.article_version_id = av.id
		WHERE av.status = $1
		GROUP BY t.name
	`

	rows, err := u.db.QueryContext(ctx, query, constanta.Published)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var name string
		var count int
		if err := rows.Scan(&name, &count); err != nil {
			return nil, err
		}
		counts[name] = count
	}

	return counts, nil
}

func (u *TagsRepo) GetTagLastUsage(ctx context.Context) (map[string]time.Time, error) {
	query := `
		SELECT t.name, MAX(av.created_at) AS last_used
		FROM tags t
		LEFT JOIN article_version_tags avt ON t.name = avt.tag_name 
		LEFT JOIN article_versions av ON avt.article_version_id = av.id
		WHERE av.status = $1
		GROUP BY t.name
	`

	rows, err := u.db.QueryContext(ctx, query, constanta.Published)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]time.Time)
	for rows.Next() {
		var name string
		var lastUsed time.Time
		if err := rows.Scan(&name, &lastUsed); err != nil {
			return nil, err
		}
		counts[name] = lastUsed
	}

	return counts, nil
}

func (u *TagsRepo) GetTagUsage(ctx context.Context) (map[string]entity.TagUsage, error) {
	query := `
		SELECT t.name, COUNT(avt.tag_name) as usage_count, MAX(av.created_at) AS last_used
		FROM tags t
		LEFT JOIN article_version_tags avt ON t.name = avt.tag_name 
		LEFT JOIN article_versions av ON avt.article_version_id = av.id
		WHERE av.status = $1
		GROUP BY t.name
	`

	rows, err := u.db.QueryContext(ctx, query, constanta.Published)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]entity.TagUsage)
	for rows.Next() {
		var name string
		var usageCount int
		var lastUsed time.Time
		if err := rows.Scan(&name, &usageCount, &lastUsed); err != nil {
			return nil, err
		}
		counts[name] = entity.TagUsage{
			Count:    usageCount,
			LastUsed: lastUsed,
		}
	}

	return counts, nil
}

const (
	getArticleTagsQuery = `SELECT 
			avt.tag_name, 
			avt.article_version_id 
		FROM 
			article_version_tags avt
		LEFT JOIN 
			article_versions av ON avt.article_version_id = av.id 
		WHERE 
			av.status = $1`
)

func (u *TagsRepo) GetArticleTags(ctx context.Context, status constanta.ArticleVersionStatus) ([]entity.ArticleVersionTag, error) {
	rows, err := u.db.QueryContext(ctx, getArticleTagsQuery, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []entity.ArticleVersionTag
	for rows.Next() {
		var tag entity.ArticleVersionTag
		if err := rows.Scan(
			&tag.TagName,
			&tag.ArticleVersionID,
		); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}
