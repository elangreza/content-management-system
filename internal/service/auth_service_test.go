package service

import (
	"context"
	"database/sql"
	"testing"

	"github.com/elangreza/content-management-system/internal/entity"
	"github.com/elangreza/content-management-system/internal/params"
	mock "github.com/elangreza/content-management-system/internal/service/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -destination=mock/mock_user_repo.go -package=service_mock . userRepo
//go:generate mockgen -destination=mock/mock_token_repo.go -package=service_mock . tokenRepo

func TestAuthService_RegisterUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		userRepo  *mock.MockuserRepo
		tokenRepo *mock.MocktokenRepo
	}

	tests := []struct {
		name    string
		prepare func(f *fields)
		input   params.RegisterUserRequest
		wantErr bool
	}{
		{
			name: "positive: user does not exist, create user success",
			prepare: func(f *fields) {
				f.userRepo.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").Return(nil, sql.ErrNoRows)
				f.userRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil)
			},
			input: params.RegisterUserRequest{
				Email:    "test@example.com",
				Password: "password",
				Name:     "Test User",
			},
			wantErr: false,
		},
		{
			name: "negative: user already exists",
			prepare: func(f *fields) {
				f.userRepo.EXPECT().GetUserByEmail(gomock.Any(), "exists@example.com").Return(&entity.User{ID: uuid.New(), Email: "exists@example.com"}, nil)
			},
			input: params.RegisterUserRequest{
				Email:    "exists@example.com",
				Password: "password",
				Name:     "Exists User",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mock.NewMockuserRepo(ctrl)
			tokenRepo := mock.NewMocktokenRepo(ctrl)
			f := &fields{userRepo, tokenRepo}
			if tt.prepare != nil {
				tt.prepare(f)
			}
			svc := &AuthService{
				UserRepo:  f.userRepo,
				TokenRepo: f.tokenRepo,
			}
			err := svc.RegisterUser(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthService_LoginUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		userRepo  *mock.MockuserRepo
		tokenRepo *mock.MocktokenRepo
	}

	tests := []struct {
		name    string
		prepare func(f *fields)
		input   params.LoginUserRequest
		wantErr bool
	}{
		{
			name: "positive: login success",
			prepare: func(f *fields) {
				user, _ := entity.NewUser("test@example.com", "password", "test")
				f.userRepo.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").Return(user, nil)
				f.userRepo.EXPECT().GetUserByEmail(gomock.Any(), gomock.Any()).AnyTimes()
				f.tokenRepo.EXPECT().GetTokenByUserID(gomock.Any(), user.ID).Return(nil, sql.ErrNoRows)
				f.tokenRepo.EXPECT().CreateToken(gomock.Any(), gomock.Any()).Return(nil)
			},
			input: params.LoginUserRequest{
				Email:    "test@example.com",
				Password: "password",
			},
			wantErr: false,
		},
		{
			name: "negative: user not found",
			prepare: func(f *fields) {
				f.userRepo.EXPECT().GetUserByEmail(gomock.Any(), "notfound@example.com").Return(nil, sql.ErrNoRows)
			},
			input: params.LoginUserRequest{
				Email:    "notfound@example.com",
				Password: "password",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mock.NewMockuserRepo(ctrl)
			tokenRepo := mock.NewMocktokenRepo(ctrl)
			f := &fields{userRepo, tokenRepo}
			if tt.prepare != nil {
				tt.prepare(f)
			}
			svc := &AuthService{
				UserRepo:  f.userRepo,
				TokenRepo: f.tokenRepo,
			}
			_, err := svc.LoginUser(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthService_ProcessToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		tokenRepo *mock.MocktokenRepo
	}

	tests := []struct {
		name    string
		prepare func(f *fields)
		input   string
		wantErr bool
	}{
		{
			name: "negative: token not found",
			prepare: func(f *fields) {
				// not valid token
			},
			input:   "notfoundtoken",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenRepo := mock.NewMocktokenRepo(ctrl)
			f := &fields{tokenRepo}
			if tt.prepare != nil {
				tt.prepare(f)
			}
			svc := &AuthService{
				TokenRepo: f.tokenRepo,
			}
			_, err := svc.ProcessToken(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthService_GetUserRoleByUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		userRepo *mock.MockuserRepo
	}

	tests := []struct {
		name    string
		prepare func(f *fields)
		input   uuid.UUID
		wantErr bool
	}{
		{
			name: "positive: user role found",
			prepare: func(f *fields) {
				userRole := &entity.UserRole{}
				f.userRepo.EXPECT().GetUserRoleByUserID(gomock.Any(), gomock.Any()).Return(userRole, nil)
			},
			input:   uuid.New(),
			wantErr: false,
		},
		{
			name: "negative: user not found",
			prepare: func(f *fields) {
				f.userRepo.EXPECT().GetUserRoleByUserID(gomock.Any(), gomock.Any()).Return(nil, sql.ErrNoRows)
			},
			input:   uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mock.NewMockuserRepo(ctrl)
			f := &fields{userRepo}
			if tt.prepare != nil {
				tt.prepare(f)
			}
			svc := &AuthService{
				UserRepo: f.userRepo,
			}
			_, err := svc.GetUserRoleByUserID(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
