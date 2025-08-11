package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/elangreza/content-management-system/internal/entity"
	"github.com/elangreza/content-management-system/internal/params"
	service_mock "github.com/elangreza/content-management-system/internal/service/mock"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -destination=mock/mock_tag_repo.go -package=service_mock . tagRepo

func TestTagService_CreateTag(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockTagRepo := service_mock.NewMocktagRepo(ctrl)
	s := &TagService{tagRepo: mockTagRepo, tagUsage: NewSafeMap[string, entity.TagUsage](), tagPairFrequency: NewSafeMap[[2]string, int](), actionTrigger: make(chan TagActionTrigger, 1)}

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "success",
			setup: func() {
				mockTagRepo.EXPECT().UpsertTags(gomock.Any(), "tag1", "tag2").Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repo error",
			setup: func() {
				mockTagRepo.EXPECT().UpsertTags(gomock.Any(), "tag1", "tag2").Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := s.CreateTag(context.Background(), "tag1", "tag2")
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTag() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTagService_GetTags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockTagRepo := service_mock.NewMocktagRepo(ctrl)
	s := &TagService{tagRepo: mockTagRepo, tagUsage: NewSafeMap[string, entity.TagUsage](), tagPairFrequency: NewSafeMap[[2]string, int](), actionTrigger: make(chan TagActionTrigger, 1)}

	now := time.Now()
	s.tagUsage.Set("tag1", entity.TagUsage{Count: 5, TrendingScore: 1.2, LastUsed: now})

	tests := []struct {
		name    string
		setup   func()
		want    []params.GetTagResponse
		wantErr bool
	}{
		{
			name: "success",
			setup: func() {
				mockTagRepo.EXPECT().GetTags(gomock.Any()).Return([]string{"tag1"}, nil)
			},
			want:    []params.GetTagResponse{{Name: "tag1", UsageCount: 5, TrendingScore: 1.2, LastUsed: now}},
			wantErr: false,
		},
		{
			name: "repo error",
			setup: func() {
				mockTagRepo.EXPECT().GetTags(gomock.Any()).Return(nil, errors.New("db error"))
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got, err := s.GetTags(context.Background(), params.GetTagsRequest{})
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTags() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagService_GetTag(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockTagRepo := service_mock.NewMocktagRepo(ctrl)
	s := &TagService{tagRepo: mockTagRepo, tagUsage: NewSafeMap[string, entity.TagUsage](), tagPairFrequency: NewSafeMap[[2]string, int](), actionTrigger: make(chan TagActionTrigger, 1)}

	now := time.Now()
	s.tagUsage.Set("tag1", entity.TagUsage{Count: 5, TrendingScore: 1.2, LastUsed: now})

	tests := []struct {
		name    string
		setup   func()
		want    *params.GetTagResponse
		wantErr bool
	}{
		{
			name: "success",
			setup: func() {
				mockTagRepo.EXPECT().GetTags(gomock.Any(), "tag1").Return([]string{"tag1"}, nil)
			},
			want:    &params.GetTagResponse{Name: "tag1", UsageCount: 5, TrendingScore: 1.2, LastUsed: now},
			wantErr: false,
		},
		{
			name: "not found",
			setup: func() {
				mockTagRepo.EXPECT().GetTags(gomock.Any(), "tag1").Return([]string{}, nil)
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got, err := s.GetTag(context.Background(), "tag1")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTag() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagService_getTagUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockTagRepo := service_mock.NewMocktagRepo(ctrl)
	s := &TagService{tagRepo: mockTagRepo, tagUsage: NewSafeMap[string, entity.TagUsage](), tagPairFrequency: NewSafeMap[[2]string, int](), actionTrigger: make(chan TagActionTrigger, 1)}

	now := time.Now()
	tagUsages := map[string]entity.TagUsage{"tag1": {Count: 3, LastUsed: now, TrendingScore: 0}}

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "success",
			setup: func() {
				mockTagRepo.EXPECT().GetTagUsage(gomock.Any()).Return(tagUsages, nil)
			},
			wantErr: false,
		},
		{
			name: "repo error",
			setup: func() {
				mockTagRepo.EXPECT().GetTagUsage(gomock.Any()).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := s.getTagUsage(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("getTagUsage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTagService_getTagPairFrequency(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockTagRepo := service_mock.NewMocktagRepo(ctrl)
	s := &TagService{tagRepo: mockTagRepo, tagUsage: NewSafeMap[string, entity.TagUsage](), tagPairFrequency: NewSafeMap[[2]string, int](), actionTrigger: make(chan TagActionTrigger, 1)}

	articleTags := []entity.ArticleVersionTag{{TagName: "tag1", ArticleVersionID: 1}, {TagName: "tag2", ArticleVersionID: 1}}

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "success",
			setup: func() {
				mockTagRepo.EXPECT().GetArticleTags(gomock.Any(), gomock.Any()).Return(articleTags, nil)
			},
			wantErr: false,
		},
		{
			name: "repo error",
			setup: func() {
				mockTagRepo.EXPECT().GetArticleTags(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := s.getTagPairFrequency(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("getTagPairFrequency() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTagService_getTagPair(t *testing.T) {
	type fields struct {
		articleRepo      articleRepo
		tagRepo          tagRepo
		tagUsage         *SafeMap[string, entity.TagUsage]
		tagPairFrequency *SafeMap[[2]string, int]
		actionTrigger    chan TagActionTrigger
	}
	type args struct {
		tags []entity.Tag
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   [][2]string
	}{
		{
			name:   "2 tags",
			fields: fields{},
			args:   args{tags: []entity.Tag{{Name: "tag1"}, {Name: "tag2"}}},
			want:   [][2]string{{"tag1", "tag2"}},
		},
		{
			name:   "3 tags",
			fields: fields{},
			args:   args{tags: []entity.Tag{{Name: "tag0"}, {Name: "tag1"}, {Name: "tag2"}}},
			want: [][2]string{
				{"tag0", "tag1"},
				{"tag0", "tag2"},
				{"tag1", "tag2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TagService{
				articleRepo:      tt.fields.articleRepo,
				tagRepo:          tt.fields.tagRepo,
				tagUsage:         tt.fields.tagUsage,
				tagPairFrequency: tt.fields.tagPairFrequency,
				actionTrigger:    tt.fields.actionTrigger,
			}
			if got := s.getTagPair(tt.args.tags); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TagService.getTagPair() = %v, want %v", got, tt.want)
			}
		})
	}
}
