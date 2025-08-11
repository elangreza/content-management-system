package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/elangreza/content-management-system/internal/constanta"
	"github.com/elangreza/content-management-system/internal/entity"
	"github.com/elangreza/content-management-system/internal/params"
	service_mock "github.com/elangreza/content-management-system/internal/service/mock"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -destination=mock/mock_article_repo.go -package=service_mock . articleRepo
//go:generate mockgen -destination=mock/mock_tag_trigger.go -package=service_mock . tagTrigger

func TestArticleService_CreateArticle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockArticleRepo := service_mock.NewMockarticleRepo(ctrl)
	mockTagTrigger := service_mock.NewMocktagTrigger(ctrl)
	service := NewArticleService(mockArticleRepo, mockTagTrigger)

	testUserID := uuid.New()
	testTags := []string{"go", "cms"}
	ctx := context.WithValue(context.Background(), constanta.LocalUserID, testUserID)

	article := entity.NewArticle("Test Title", "Test Body", testUserID)
	articleVersion := entity.NewArticleVersion(article.ID, "Test Title", "Test Body", testUserID, 1, testTags)

	tests := []struct {
		name    string
		prepare func()
		ctx     context.Context
		input   params.CreateArticleRequest
		wantErr bool
	}{
		{
			name: "success",
			prepare: func() {
				mockArticleRepo.EXPECT().CreateArticle(gomock.Any(), *article, *articleVersion).Return(int64(1), int64(2), nil)
				mockTagTrigger.EXPECT().CreateTagTrigger(gomock.Any(), gomock.Any())
			},
			ctx: ctx,
			input: params.CreateArticleRequest{
				Title: "Test Title",
				Body:  "Test Body",
				Tags:  testTags,
			},
			wantErr: false,
		},
		{
			name: "error from repo",
			prepare: func() {
				mockArticleRepo.EXPECT().CreateArticle(gomock.Any(), *article, *articleVersion).Return(int64(0), int64(0), errors.New("repo error"))
			},
			ctx: ctx,
			input: params.CreateArticleRequest{
				Title: "Test Title",
				Body:  "Test Body",
				Tags:  testTags,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare()
			_, err := service.CreateArticle(tt.ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateArticle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestArticleService_DeleteArticle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockArticleRepo := service_mock.NewMockarticleRepo(ctrl)
	mockTagTrigger := service_mock.NewMocktagTrigger(ctrl)
	service := NewArticleService(mockArticleRepo, mockTagTrigger)

	tests := []struct {
		name    string
		prepare func()
		input   int64
		wantErr bool
	}{
		{
			name: "success",
			prepare: func() {
				mockArticleRepo.EXPECT().DeleteArticle(gomock.Any(), int64(1)).Return(nil)
				mockTagTrigger.EXPECT().CreateTagTrigger(gomock.Any(), gomock.Any())
			},
			input:   1,
			wantErr: false,
		},
		{
			name: "repo error",
			prepare: func() {
				mockArticleRepo.EXPECT().DeleteArticle(gomock.Any(), int64(2)).Return(errors.New("repo error"))
			},
			input:   2,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare()
			err := service.DeleteArticle(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteArticle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestArticleService_UpdateStatusArticle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockArticleRepo := service_mock.NewMockarticleRepo(ctrl)
	mockTagTrigger := service_mock.NewMocktagTrigger(ctrl)
	service := NewArticleService(mockArticleRepo, mockTagTrigger)

	testUserID := uuid.New()
	ctx := context.WithValue(context.Background(), constanta.LocalUserID, testUserID)

	articleVersion := &entity.ArticleVersion{Status: constanta.Draft}
	articleTags := []entity.Tag{{Name: "go"}}

	tests := []struct {
		name    string
		prepare func()
		ctx     context.Context
		wantErr bool
	}{
		{
			name: "success",
			prepare: func() {
				mockArticleRepo.EXPECT().GetArticleVersionWithIDAndArticleID(gomock.Any(), int64(1), int64(1)).Return(articleVersion, nil)
				mockArticleRepo.EXPECT().GetTagsWithArticleVersionID(gomock.Any(), int64(1)).Return(articleTags, nil)
				mockTagTrigger.EXPECT().CreateTagTrigger(gomock.Any(), gomock.Any())
				mockArticleRepo.EXPECT().UpdateArticleStatus(gomock.Any(), int64(1), int64(1), constanta.Published, constanta.Draft, testUserID).Return(nil)
			},
			ctx:     ctx,
			wantErr: false,
		},
		{
			name: "not found",
			prepare: func() {
				mockArticleRepo.EXPECT().GetArticleVersionWithIDAndArticleID(gomock.Any(), int64(1), int64(1)).Return(nil, sql.ErrNoRows)
			},
			ctx:     ctx,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare()
			err := service.UpdateStatusArticle(tt.ctx, int64(1), int64(1), constanta.Published)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateStatusArticle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestArticleService_CreateArticleVersionWithReferenceFromArticleID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockArticleRepo := service_mock.NewMockarticleRepo(ctrl)
	mockTagTrigger := service_mock.NewMocktagTrigger(ctrl)
	service := NewArticleService(mockArticleRepo, mockTagTrigger)

	testUserID := uuid.New()
	testTags := []string{"go", "cms"}
	ctx := context.WithValue(context.Background(), constanta.LocalUserID, testUserID)
	article := &entity.Article{ID: 1, VersionSequence: 1}
	// articleVersion := &entity.ArticleVersion{Title: "v2", Body: "b2"}

	tests := []struct {
		name    string
		prepare func()
		ctx     context.Context
		inputID int64
		input   params.CreateArticleVersionRequest
		wantErr bool
	}{
		{
			name: "success",
			prepare: func() {
				mockArticleRepo.EXPECT().GetArticleWithID(gomock.Any(), int64(1)).Return(article, nil)
				mockArticleRepo.EXPECT().CreateArticleVersion(gomock.Any(), gomock.Any()).Return(int64(2), nil)
				mockTagTrigger.EXPECT().CreateTagTrigger(gomock.Any(), gomock.Any())
			},
			ctx:     ctx,
			inputID: 1,
			input:   params.CreateArticleVersionRequest{Title: "v2", Body: "b2", Tags: testTags},
			wantErr: false,
		},
		{
			name: "not found",
			prepare: func() {
				mockArticleRepo.EXPECT().GetArticleWithID(gomock.Any(), int64(2)).Return(nil, sql.ErrNoRows)
			},
			ctx:     ctx,
			inputID: 2,
			input:   params.CreateArticleVersionRequest{Title: "v2", Body: "b2", Tags: testTags},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare()
			_, err := service.CreateArticleVersionWithReferenceFromArticleID(tt.ctx, tt.inputID, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateArticleVersionWithReferenceFromArticleID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestArticleService_CreateArticleVersionWithReferenceFromArticleIDAindVersionID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockArticleRepo := service_mock.NewMockarticleRepo(ctrl)
	mockTagTrigger := service_mock.NewMocktagTrigger(ctrl)
	service := NewArticleService(mockArticleRepo, mockTagTrigger)

	testUserID := uuid.New()
	testTags := []string{"go", "cms"}
	ctx := context.WithValue(context.Background(), constanta.LocalUserID, testUserID)
	article := &entity.Article{ID: 1, VersionSequence: 1}
	articleVersion := &entity.ArticleVersion{Title: "v2", Body: "b2"}

	tests := []struct {
		name       string
		prepare    func()
		ctx        context.Context
		inputID    int64
		inputVerID int64
		input      params.CreateArticleVersionRequest
		wantErr    bool
	}{
		{
			name: "success",
			prepare: func() {
				mockArticleRepo.EXPECT().GetArticleWithID(gomock.Any(), int64(1)).Return(article, nil)
				mockArticleRepo.EXPECT().GetArticleVersionWithIDAndArticleID(gomock.Any(), int64(1), int64(1)).Return(articleVersion, nil)
				mockArticleRepo.EXPECT().GetTagsWithArticleVersionID(gomock.Any(), int64(1)).Return([]entity.Tag{}, nil)
				mockArticleRepo.EXPECT().CreateArticleVersion(gomock.Any(), gomock.Any()).Return(int64(2), nil)
				mockTagTrigger.EXPECT().CreateTagTrigger(gomock.Any(), gomock.Any())
			},
			ctx:        ctx,
			inputID:    1,
			inputVerID: 1,
			input:      params.CreateArticleVersionRequest{Title: "v2", Body: "b2", Tags: testTags},
			wantErr:    false,
		},
		{
			name: "not found",
			prepare: func() {
				mockArticleRepo.EXPECT().GetArticleWithID(gomock.Any(), int64(2)).Return(nil, sql.ErrNoRows)
			},
			ctx:        ctx,
			inputID:    2,
			inputVerID: 2,
			input:      params.CreateArticleVersionRequest{Title: "v2", Body: "b2", Tags: testTags},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare()
			_, err := service.CreateArticleVersionWithReferenceFromArticleIDAindVersionID(tt.ctx, tt.inputID, tt.inputVerID, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateArticleVersionWithReferenceFromArticleIDAindVersionID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestArticleService_GetArticleWithID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockArticleRepo := service_mock.NewMockarticleRepo(ctrl)
	mockTagTrigger := service_mock.NewMocktagTrigger(ctrl)
	service := NewArticleService(mockArticleRepo, mockTagTrigger)

	article := &entity.Article{ID: 1, DraftedVersionID: 1, PublishedVersionID: 1, ArchivedVersionID: 1}
	version := &entity.ArticleVersion{ArticleID: 1, ArticleVersionID: 1, Title: "t", Body: "b", Version: 1, Status: constanta.Draft, Tags: []entity.Tag{{Name: "go"}}}
	tags := []entity.Tag{{Name: "go"}}
	ctx := context.WithValue(context.Background(), constanta.LocalUserCanReadDraftedAndArchivedArticle, false)

	tests := []struct {
		name    string
		prepare func()
		ctx     context.Context
		inputID int64
		wantErr bool
	}{
		{
			name: "success",
			prepare: func() {
				mockArticleRepo.EXPECT().GetArticleWithID(gomock.Any(), int64(1)).Return(article, nil)
				mockArticleRepo.EXPECT().GetArticleVersionWithIDAndArticleID(gomock.Any(), int64(1), int64(1)).Return(version, nil)
				mockArticleRepo.EXPECT().GetTagsWithArticleVersionID(gomock.Any(), int64(1)).Return(tags, nil)
			},
			ctx:     ctx,
			inputID: 1,
			wantErr: false,
		},
		{
			name: "not found",
			prepare: func() {
				mockArticleRepo.EXPECT().GetArticleWithID(gomock.Any(), int64(2)).Return(nil, sql.ErrNoRows)
			},
			ctx:     ctx,
			inputID: 2,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare()
			_, err := service.GetArticleWithID(tt.ctx, tt.inputID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetArticleWithID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestArticleService_GetArticleVersionWithIDAndArticleID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockArticleRepo := service_mock.NewMockarticleRepo(ctrl)
	mockTagTrigger := service_mock.NewMocktagTrigger(ctrl)
	service := NewArticleService(mockArticleRepo, mockTagTrigger)

	version := &entity.ArticleVersion{ArticleID: 1, ArticleVersionID: 1, Title: "t", Body: "b", Version: 1, Status: constanta.Draft, Tags: []entity.Tag{{Name: "go"}}}
	ctx := context.WithValue(context.Background(), constanta.LocalUserCanReadDraftedAndArchivedArticle, true)

	tests := []struct {
		name       string
		prepare    func()
		ctx        context.Context
		inputID    int64
		inputVerID int64
		wantErr    bool
	}{
		{
			name: "success",
			prepare: func() {
				mockArticleRepo.EXPECT().GetArticleVersionWithIDAndArticleID(gomock.Any(), int64(1), int64(1)).Return(version, nil)
			},
			ctx:        ctx,
			inputID:    1,
			inputVerID: 1,
			wantErr:    false,
		},
		{
			name: "not found",
			prepare: func() {
				mockArticleRepo.EXPECT().GetArticleVersionWithIDAndArticleID(gomock.Any(), int64(2), int64(2)).Return(nil, sql.ErrNoRows)
			},
			ctx:        ctx,
			inputID:    2,
			inputVerID: 2,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare()
			_, err := service.GetArticleVersionWithIDAndArticleID(tt.ctx, tt.inputID, tt.inputVerID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetArticleVersionWithIDAndArticleID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestArticleService_GetArticleVersions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockArticleRepo := service_mock.NewMockarticleRepo(ctrl)
	mockTagTrigger := service_mock.NewMocktagTrigger(ctrl)
	service := NewArticleService(mockArticleRepo, mockTagTrigger)

	ctx := context.WithValue(context.Background(), constanta.LocalUserCanReadDraftedAndArchivedArticle, true)
	articleVersions := []entity.ArticleVersion{{ArticleID: 1, ArticleVersionID: 1, Title: "t", Body: "b", Version: 1, Status: constanta.Draft}}

	tests := []struct {
		name    string
		prepare func()
		ctx     context.Context
		inputID int64
		wantErr bool
	}{
		{
			name: "success",
			prepare: func() {
				mockArticleRepo.EXPECT().GetArticleVersionsWithArticleIDAndStatuses(gomock.Any(), int64(1), gomock.Any()).Return(articleVersions, nil)
			},
			ctx:     ctx,
			inputID: 1,
			wantErr: false,
		},
		{
			name: "repo error",
			prepare: func() {
				mockArticleRepo.EXPECT().GetArticleVersionsWithArticleIDAndStatuses(gomock.Any(), int64(2), gomock.Any()).Return(nil, errors.New("repo error"))
			},
			ctx:     ctx,
			inputID: 2,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare()
			_, err := service.GetArticleVersions(tt.ctx, tt.inputID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetArticleVersions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestArticleService_GetArticles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockArticleRepo := service_mock.NewMockarticleRepo(ctrl)
	mockTagTrigger := service_mock.NewMocktagTrigger(ctrl)
	service := NewArticleService(mockArticleRepo, mockTagTrigger)

	ctx := context.WithValue(context.Background(), constanta.LocalUserCanReadDraftedAndArchivedArticle, true)
	query := params.GetArticlesQueryParams{Search: "test"}
	articleVersions := []entity.ArticleVersion{{ArticleID: 1, ArticleVersionID: 1, Title: "t", Body: "b", Version: 1, Status: constanta.Draft}}

	tests := []struct {
		name    string
		prepare func()
		ctx     context.Context
		input   params.GetArticlesQueryParams
		wantErr bool
	}{
		{
			name: "success",
			prepare: func() {
				mockArticleRepo.EXPECT().GetArticles(gomock.Any(), gomock.Any()).Return(articleVersions, nil)
			},
			ctx:     ctx,
			input:   query,
			wantErr: false,
		},
		{
			name: "repo error",
			prepare: func() {
				mockArticleRepo.EXPECT().GetArticles(gomock.Any(), gomock.Any()).Return(nil, errors.New("repo error"))
			},
			ctx:     ctx,
			input:   query,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare()
			_, err := service.GetArticles(tt.ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetArticles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
