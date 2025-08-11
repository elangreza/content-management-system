package service_test

import (
	"context"
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/elangreza/content-management-system/internal/constanta"
	"github.com/elangreza/content-management-system/internal/entity"
	"github.com/elangreza/content-management-system/internal/params"
	"github.com/elangreza/content-management-system/internal/service"
	service_mock "github.com/elangreza/content-management-system/internal/service/mock"
	"github.com/elangreza/content-management-system/internal/sharevar"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestProfileService_GetUserProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := service_mock.NewMockuserRepo(ctrl)
	ps := service.NewProfileService(mockUserRepo)

	testUserID := uuid.New()
	testTime := time.Now()

	tests := []struct {
		name      string
		ctx       context.Context
		mockSetup func()
		want      *params.UserProfileResponse
		wantErr   bool
	}{
		{
			name: "positive case: user found",
			ctx: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, constanta.LocalUserID, testUserID)
				return ctx
			}(),
			mockSetup: func() {
				mockUserRepo.EXPECT().GetUserByID(gomock.Any(), testUserID).Return(&entity.User{
					ID:        testUserID,
					Name:      "John Doe",
					Email:     "john@example.com",
					Role:      sharevar.ContentWriter,
					CreatedAt: testTime,
					UpdatedAt: sql.NullTime{Time: testTime, Valid: true},
				}, nil)
			},
			want: &params.UserProfileResponse{
				Name:      "John Doe",
				Email:     "john@example.com",
				Role:      sharevar.ContentWriter.GetValue(),
				RoleName:  sharevar.ContentWriter.GetName(),
				CreatedAt: testTime,
				UpdatedAt: &testTime,
			},
			wantErr: false,
		},
		{
			name:      "negative case: userID not in context",
			ctx:       context.Background(),
			mockSetup: func() {},
			want:      nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			got, err := ps.GetUserProfile(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUserProfile() = %v, want %v", got, tt.want)
			}
		})
	}
}

// sqlNullTime returns a sql.NullTime with Valid true
func sqlNullTime(t time.Time) sql.NullTime {
	return sql.NullTime{
		Valid: true,
		Time:  t,
	}
}
