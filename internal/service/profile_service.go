package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/elangreza/content-management-system/internal/constanta"
	"github.com/elangreza/content-management-system/internal/params"
	"github.com/elangreza/content-management-system/internal/sharevar"
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
	userID, ok := ctx.Value(constanta.LocalUserID).(uuid.UUID)
	if !ok {
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

	fmt.Println(
		sharevar.ContentWriter.GetValue())
	fmt.Println(
		sharevar.Editor.GetValue())

	roleName := "custom role"
	switch user.Role.GetValue() {
	case sharevar.ContentWriter.GetValue():
		roleName = sharevar.ContentWriter.GetName()
	case sharevar.Editor.GetValue():
		roleName = sharevar.Editor.GetName()
	}

	return &params.UserProfileResponse{
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role.GetValue(),
		RoleName:  roleName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: updateAt,
	}, nil

}
