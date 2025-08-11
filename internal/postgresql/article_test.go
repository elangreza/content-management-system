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
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestArticleRepo_CreateArticle(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "positive case - create article successfully",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(createArticleQuery)).WithArgs(uuid.Nil).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
				m.ExpectQuery(regexp.QuoteMeta(createArticleVersionQuery)).WithArgs(int64(1), "title", "body", int64(1), constanta.Published, uuid.Nil).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(2)))
				m.ExpectExec(regexp.QuoteMeta(updateLatestArticleVersionQuery)).WithArgs(uuid.Nil, int64(2), int64(1), int64(1)).WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "negative case - create article fails on article insert",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(createArticleQuery)).WithArgs(uuid.Nil).WillReturnError(errors.New("insert error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
			assert.NoError(t, err)
			defer db.Close()
			repo := NewArticleRepo(db)
			article := entity.Article{CreatedBy: uuid.Nil, VersionSequence: 1}
			articleVersion := entity.ArticleVersion{Title: "title", Body: "body", Version: 1, Status: constanta.Published, CreatedBy: uuid.Nil}
			tt.mock(mock)
			_, _, err = repo.CreateArticle(context.Background(), article, articleVersion)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestArticleRepo_DeleteArticle(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "positive case - delete article successfully",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectExec(regexp.QuoteMeta(resetArticlePublishedAndDraftedToNullQuery)).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(regexp.QuoteMeta(deleteArticleVersionsByArticleIdQuery)).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(regexp.QuoteMeta(deleteArticleByArticleIdQuery)).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "negative case - delete article fails on first exec",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectExec(regexp.QuoteMeta(resetArticlePublishedAndDraftedToNullQuery)).WithArgs(int64(1)).WillReturnError(errors.New("delete error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()
			repo := NewArticleRepo(db)
			tt.mock(mock)
			err = repo.DeleteArticle(context.Background(), 1)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestArticleRepo_GetArticleVersionWithIDAndArticleID(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(sqlmock.Sqlmock)
		want    *entity.ArticleVersion
		wantErr bool
	}{
		{
			name: "positive case - get article version successfully",
			mock: func(m sqlmock.Sqlmock) {
				row := sqlmock.NewRows([]string{"id", "article_id", "title", "body", "version", "status", "tag_relationship_score", "created_by", "created_at", "updated_by", "updated_at"}).
					AddRow(int64(2), int64(1), "title", "body", int64(1), constanta.Published, 0.0, uuid.Nil, time.Now(), uuid.Nil, nil)
				m.ExpectQuery(regexp.QuoteMeta(getArticleVersionWithIDAndArticleIDQuery)).WithArgs(int64(1), int64(2)).WillReturnRows(row)
			},
			want:    &entity.ArticleVersion{ArticleVersionID: 2, ArticleID: 1, Title: "title", Body: "body", Version: 1, Status: constanta.Published, TagRelationShipScore: 0.0, CreatedBy: uuid.Nil},
			wantErr: false,
		},
		{
			name: "negative case - query returns error",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(getArticleVersionWithIDAndArticleIDQuery)).WithArgs(int64(1), int64(2)).WillReturnError(errors.New("query error"))
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
			repo := NewArticleRepo(db)
			tt.mock(mock)
			got, err := repo.GetArticleVersionWithIDAndArticleID(context.Background(), 1, 2)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.ArticleVersionID, got.ArticleVersionID)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestArticleRepo_UpdateArticleStatus(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "positive case - update article status successfully",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectExec(regexp.QuoteMeta(updateArticleVersionQuery)).WithArgs(constanta.Archived, uuid.Nil, int64(1), int64(2)).WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(regexp.QuoteMeta(updateArticleArchivedIdQuery)).WithArgs(int64(2), uuid.Nil, int64(1)).WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(regexp.QuoteMeta(updateArticlePublishedIdQuery)).WithArgs(nil, uuid.Nil, int64(1)).WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "negative case - update article status fails",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectExec(regexp.QuoteMeta(updateArticleVersionQuery)).WithArgs(constanta.Archived, uuid.Nil, int64(1), int64(2)).WillReturnError(errors.New("update error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
			assert.NoError(t, err)
			defer db.Close()
			repo := NewArticleRepo(db)
			tt.mock(mock)
			err = repo.UpdateArticleStatus(context.Background(), 1, 2, constanta.Archived, constanta.Published, uuid.Nil)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestArticleRepo_CreateArticleVersion(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "positive case - create article version successfully",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(createArticleVersionQuery)).WithArgs(int64(1), "title", "body", int64(1), constanta.Published, uuid.Nil).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(2)))
				m.ExpectExec(regexp.QuoteMeta(updateLatestArticleVersionQuery)).WithArgs(uuid.Nil, int64(2), int64(1), int64(1)).WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "negative case - create article version fails",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(createArticleVersionQuery)).WithArgs(int64(1), "title", "body", int64(1), constanta.Published, uuid.Nil).WillReturnError(errors.New("insert error"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()
			repo := NewArticleRepo(db)
			articleVersion := entity.ArticleVersion{ArticleID: 1, Title: "title", Body: "body", Version: 1, Status: constanta.Published, CreatedBy: uuid.Nil}
			tt.mock(mock)
			_, err = repo.CreateArticleVersion(context.Background(), articleVersion)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestArticleRepo_GetArticleWithID(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(sqlmock.Sqlmock)
		want    *entity.Article
		wantErr bool
	}{
		{
			name: "positive case - get article with id successfully",
			mock: func(m sqlmock.Sqlmock) {
				row := sqlmock.NewRows([]string{"id", "published_version_id", "drafted_version_id", "archived_version_id", "version_sequence", "created_by", "created_at", "updated_by", "updated_at"}).
					AddRow(int64(1), int64(2), int64(3), int64(4), int64(1), uuid.Nil, time.Now(), uuid.Nil, time.Now())
				m.ExpectQuery(regexp.QuoteMeta(getArticleWithIDQuery)).WithArgs(int64(1)).WillReturnRows(row)
			},
			want:    &entity.Article{ID: 1, PublishedVersionID: 2, DraftedVersionID: 3, ArchivedVersionID: 4, VersionSequence: 1, CreatedBy: uuid.Nil},
			wantErr: false,
		},
		{
			name: "negative case - query returns error",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(getArticleWithIDQuery)).WithArgs(int64(1)).WillReturnError(errors.New("query error"))
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
			repo := NewArticleRepo(db)
			tt.mock(mock)
			got, err := repo.GetArticleWithID(context.Background(), 1)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.ID, got.ID)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestArticleRepo_GetArticleVersionsWithArticleIDAndStatuses(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(sqlmock.Sqlmock)
		want    []entity.ArticleVersion
		wantErr bool
	}{
		{
			name: "positive case - get article versions successfully",
			mock: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "article_id", "title", "body", "version", "status", "tag_relationship_score", "created_by", "created_at", "updated_by", "updated_at"}).
					AddRow(int64(2), int64(1), "title", "body", int64(1), constanta.Published, 0.0, uuid.Nil, time.Now(), uuid.Nil, time.Now())
				m.ExpectQuery(regexp.QuoteMeta(getArticleVersionsWithArticleIDAndStatusesQuery)).WithArgs(int64(1), pq.Array([]constanta.ArticleVersionStatus{constanta.Published})).WillReturnRows(rows)
			},
			want:    []entity.ArticleVersion{{ArticleVersionID: 2, ArticleID: 1, Title: "title", Body: "body", Version: 1, Status: constanta.Published, TagRelationShipScore: 0.0, CreatedBy: uuid.Nil}},
			wantErr: false,
		},
		{
			name: "negative case - query returns error",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(getArticleVersionsWithArticleIDAndStatusesQuery)).WithArgs(int64(1), pq.Array([]constanta.ArticleVersionStatus{constanta.Published})).WillReturnError(errors.New("query error"))
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
			repo := NewArticleRepo(db)
			tt.mock(mock)
			got, err := repo.GetArticleVersionsWithArticleIDAndStatuses(context.Background(), 1, constanta.Published)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want[0].ArticleVersionID, got[0].ArticleVersionID)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestArticleRepo_GetTagsWithArticleVersionID(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(sqlmock.Sqlmock)
		want    []entity.Tag
		wantErr bool
	}{
		{
			name: "positive case - get tags with article version id successfully",
			mock: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"tag_name"}).AddRow("go").AddRow("test")
				m.ExpectQuery(regexp.QuoteMeta(getTagsWithArticleVersionIDQuery)).WithArgs(int64(2)).WillReturnRows(rows)
			},
			want:    []entity.Tag{{Name: "go"}, {Name: "test"}},
			wantErr: false,
		},
		{
			name: "negative case - query returns error",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(getTagsWithArticleVersionIDQuery)).WithArgs(int64(2)).WillReturnError(errors.New("query error"))
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
			repo := NewArticleRepo(db)
			tt.mock(mock)
			got, err := repo.GetTagsWithArticleVersionID(context.Background(), 2)
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

func TestArticleRepo_UpdateArticleVersionRelationshipScore(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "positive case - update relationship score successfully",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(regexp.QuoteMeta(UpdateArticleVersionRelationshipScoreQuery)).WithArgs(0.5, int64(2)).WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "negative case - update relationship score fails",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(regexp.QuoteMeta(UpdateArticleVersionRelationshipScoreQuery)).WithArgs(0.5, int64(2)).WillReturnError(errors.New("update error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()
			repo := NewArticleRepo(db)
			tt.mock(mock)
			err = repo.UpdateArticleVersionRelationshipScore(context.Background(), 2, 0.5)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// Note: For brevity, GetArticles is omitted here, as it requires more complex query and mock setup.
// You can request it specifically if needed.
