package postgresql

import (
	"context"
	"database/sql"

	"github.com/elangreza/content-management-system/internal/entity"
	"github.com/google/uuid"
)

type (
	UserRepo struct {
		db *sql.DB
	}
)

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

const (
	createUserQuery = `INSERT INTO users
	(id, "name", email, "password", "role")
	VALUES($1, $2, $3, $4, $5);`
)

// CreateUser implements userRepo.
func (u *UserRepo) CreateUser(ctx context.Context, user entity.User) error {
	_, err := u.db.ExecContext(ctx, createUserQuery,
		user.ID,
		user.Name,
		user.Email,
		user.GetPassword(),
		user.Role)
	if err != nil {
		return err
	}

	return nil
}

const (
	getUserByEmailQuery = `SELECT 
		id, 
		"name", 
		email, 
		"password", 
		"role", 
		created_at,
		updated_at
	FROM 
		users
	WHERE 
		email=$1;`
)

// GetUserByEmail implements userRepo.
func (u *UserRepo) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {

	user := &entity.User{}
	password := []byte{}
	err := u.db.QueryRowContext(ctx, getUserByEmailQuery, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	user.SetPassword(password)

	return user, nil
}

const (
	getUserByIDQuery = `SELECT 
		id, 
		"name", 
		email, 
		"password", 
		"role", 
		created_at,
		updated_at
	FROM 
		users
	WHERE 
		id=$1;`
)

// GetUserByID implements userRepo.
func (u *UserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user := &entity.User{}
	password := []byte{}
	err := u.db.QueryRowContext(ctx, getUserByIDQuery, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	user.SetPassword(password)

	return user, nil
}

const (
	getUserRoleByUserIDQuery = `SELECT 
		role
	FROM		
		users
	WHERE 
		id=$1;`
)

// GetUserRoleByUserID implements userRepo.
func (u *UserRepo) GetUserRoleByUserID(ctx context.Context, id uuid.UUID) (*entity.UserRole, error) {
	userRole := &entity.UserRole{}
	err := u.db.QueryRowContext(ctx, getUserRoleByUserIDQuery, id).Scan(
		&userRole,
	)
	if err != nil {
		return nil, err
	}

	return userRole, nil
}
