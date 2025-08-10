package postgresql

import (
	"context"
	"database/sql"
	"time"

	"github.com/elangreza/content-management-system/internal/constanta"
	"github.com/elangreza/content-management-system/internal/entity"
	"github.com/lib/pq"
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

const (
	getTagsQuery = `SELECT name FROM tags WHERE (name IN ($1) OR $1 IS NULL)`
)

func (u *TagsRepo) GetTags(ctx context.Context, names ...string) ([]string, error) {
	rows, err := u.db.QueryContext(ctx, getTagsQuery, pq.Array(names))
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

const getTagUsageQuery = `
	SELECT t.name, COUNT(avt.tag_name) as usage_count, MAX(av.created_at) AS last_used
	FROM tags t
	LEFT JOIN article_version_tags avt ON t.name = avt.tag_name 
	LEFT JOIN article_versions av ON avt.article_version_id = av.id
	WHERE av.status = $1
	GROUP BY t.name
`

func (u *TagsRepo) GetTagUsage(ctx context.Context) (map[string]entity.TagUsage, error) {

	rows, err := u.db.QueryContext(ctx, getTagUsageQuery, constanta.Published)
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
