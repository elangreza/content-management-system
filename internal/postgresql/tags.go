package postgresql

import (
	"context"
	"database/sql"
)

type (
	TagsRepo struct {
		db *sql.DB
	}
)

func NewTagsRepo(db *sql.DB) *TagsRepo {
	return &TagsRepo{
		db: db,
	}
}

const (
	upsertTagQuery = `INSERT INTO tags ("name") VALUES ($1) ON CONFLICT (name) DO NOTHING`
)

// UpsertTags implements TagsRepo.
func (u *TagsRepo) UpsertTags(ctx context.Context, name string) error {
	_, err := u.db.ExecContext(ctx, upsertTagQuery, name)
	return err
}
