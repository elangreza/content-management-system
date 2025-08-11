package postgresql

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/elangreza/content-management-system/internal/constanta"
	"github.com/elangreza/content-management-system/internal/entity"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestTagsRepo_UpsertTags(t *testing.T) {
	type args struct {
		names []string
	}
	tests := []struct {
		name    string
		args    args
		mock    func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "positive case - upsert tags successfully",
			args: args{names: []string{"go", "test"}},
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectPrepare(regexp.QuoteMeta(upsertTagQuery)).
					ExpectExec().
					WithArgs("go").
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(regexp.QuoteMeta(upsertTagQuery)).
					WithArgs("test").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "negative case - prepare returns error",
			args: args{names: []string{"fail"}},
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectPrepare(regexp.QuoteMeta(upsertTagQuery)).
					WillReturnError(errors.New("prepare error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()
			tt.mock(mock)

			repo := NewTagRepo(db)
			err = repo.UpsertTags(context.Background(), tt.args.names...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTagsRepo_GetTags(t *testing.T) {
	type args struct {
		names []string
	}
	tests := []struct {
		name    string
		args    args
		mock    func(sqlmock.Sqlmock)
		want    []string
		wantErr bool
	}{
		{
			name: "positive case - get tags successfully",
			args: args{names: []string{"go", "test"}},
			mock: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"name"}).
					AddRow("go").
					AddRow("test")
				m.ExpectQuery(regexp.QuoteMeta(getTagsQuery)).
					WithArgs(pq.Array([]string{"go", "test"})).
					WillReturnRows(rows)
			},
			want:    []string{"go", "test"},
			wantErr: false,
		},
		{
			name: "negative case - query returns error",
			args: args{names: []string{"fail"}},
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(getTagsQuery)).
					WithArgs(pq.Array([]string{"fail"})).
					WillReturnError(errors.New("query error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()
			tt.mock(mock)

			repo := NewTagRepo(db)
			got, err := repo.GetTags(context.Background(), tt.args.names...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTagsRepo_GetTagUsage(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(sqlmock.Sqlmock)
		want    map[string]entity.TagUsage
		wantErr bool
	}{
		{
			name: "positive case - get tag usage successfully",
			mock: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"name", "usage_count", "last_used"}).
					AddRow("go", 2, time.Date(2025, 8, 11, 10, 0, 0, 0, time.UTC)).
					AddRow("test", 1, time.Date(2025, 8, 10, 9, 0, 0, 0, time.UTC))
				m.ExpectQuery(regexp.QuoteMeta(getTagUsageQuery)).
					WithArgs(constanta.Published).
					WillReturnRows(rows)
			},
			want: map[string]entity.TagUsage{
				"go": {
					Count:    2,
					LastUsed: time.Date(2025, 8, 11, 10, 0, 0, 0, time.UTC),
				},
				"test": {
					Count:    1,
					LastUsed: time.Date(2025, 8, 10, 9, 0, 0, 0, time.UTC),
				},
			},
			wantErr: false,
		},
		{
			name: "negative case - query returns error",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(getTagUsageQuery)).
					WithArgs(constanta.Published).
					WillReturnError(errors.New("query error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()
			tt.mock(mock)

			repo := NewTagRepo(db)
			got, err := repo.GetTagUsage(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTagsRepo_GetArticleTags(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(sqlmock.Sqlmock)
		status  constanta.ArticleVersionStatus
		want    []entity.ArticleVersionTag
		wantErr bool
	}{
		{
			name:   "positive case - get article tags successfully",
			status: constanta.Published,
			mock: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"tag_name", "article_version_id"}).
					AddRow("go", int64(1)).
					AddRow("test", int64(2))
				m.ExpectQuery(regexp.QuoteMeta(getArticleTagsQuery)).
					WithArgs(constanta.Published).
					WillReturnRows(rows)
			},
			want: []entity.ArticleVersionTag{
				{TagName: "go", ArticleVersionID: 1},
				{TagName: "test", ArticleVersionID: 2},
			},
			wantErr: false,
		},
		{
			name:   "negative case - query returns error",
			status: constanta.Published,
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(getArticleTagsQuery)).
					WithArgs(constanta.Published).
					WillReturnError(errors.New("query error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()
			tt.mock(mock)

			repo := NewTagRepo(db)
			got, err := repo.GetArticleTags(context.Background(), tt.status)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
