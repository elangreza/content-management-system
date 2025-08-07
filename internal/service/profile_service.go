package service

import (
	"context"
	"errors"
	"time"

	"github.com/elangreza/content-management-system/internal/constanta"
	"github.com/elangreza/content-management-system/internal/params"
	"github.com/google/uuid"
)

type (
	ProfileService struct {
		UserRepo userRepo
	}
)

func NewProfileService(userRepo userRepo) *ProfileService {
	return &ProfileService{UserRepo: userRepo}
}

func (ps *ProfileService) GetUserProfile(ctx context.Context) (*params.UserProfileResponse, error) {
	localUserID, ok := ctx.Value(constanta.LocalUserID).(string)
	if !ok {
		return nil, errors.New("error when handle ctx value")
	}
	userID, err := uuid.Parse(localUserID)
	if err != nil {
		return nil, errors.New("error when parsing userID")
	}

	user, err := ps.UserRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var updateAt *time.Time
	if user.UpdatedAt.Valid {
		updateAt = &user.UpdatedAt.Time
	}

	return &params.UserProfileResponse{
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: updateAt,
	}, nil

}
