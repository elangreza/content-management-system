package postgresql

import (
	"context"
	"database/sql"

	"github.com/elangreza/content-management-system/internal/entity"
	"github.com/google/uuid"
)

type (
	TokenRepo struct {
		db *sql.DB
	}
)

func NewTokenRepo(db *sql.DB) *TokenRepo {
	return &TokenRepo{
		db: db,
	}
}

const (
	createTokenQuery = `INSERT INTO public.tokens
(id, user_id, "token", token_type, issued_at, expired_at, duration)
VALUES($1, $2, $3, $4, $5, $6, $7);`
)

// CreateToken implements tokenRepo.
func (u *TokenRepo) CreateToken(ctx context.Context, token entity.Token) error {
	_, err := u.db.ExecContext(ctx, createTokenQuery,
		token.ID,
		token.UserID,
		token.Token,
		token.TokenType,
		token.IssuedAt,
		token.ExpiredAt,
		token.Duration,
	)
	if err != nil {
		return err
	}

	return nil
}

const (
	getTokenByTokenIDQuery = `SELECT 
		id, 
		user_id, 
		"token", 
		token_type, 
		issued_at, 
		expired_at, 
		duration, 
		created_at, 
		updated_at
	FROM tokens
	WHERE id = $1
	;`
)

// GetTokenByTokenID implements tokenRepo.
func (u *TokenRepo) GetTokenByTokenID(ctx context.Context, tokenID uuid.UUID) (*entity.Token, error) {
	token := &entity.Token{}
	err := u.db.QueryRowContext(ctx, getTokenByTokenIDQuery, tokenID).Scan(
		&token.ID,
		&token.UserID,
		&token.Token,
		&token.TokenType,
		&token.IssuedAt,
		&token.ExpiredAt,
		&token.Duration,
		&token.CreatedAt,
		&token.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return token, nil
}

const (
	getTokenByUserIDQuery = `SELECT 
		id, 
		user_id, 
		"token", 
		token_type, 
		issued_at, 
		expired_at, 
		duration, 
		created_at, 
		updated_at
	FROM tokens
	WHERE user_id = $1
	;`
)

// GetTokenByUserID implements tokenRepo.
func (u *TokenRepo) GetTokenByUserID(ctx context.Context, userID uuid.UUID) (*entity.Token, error) {
	token := &entity.Token{}
	err := u.db.QueryRowContext(ctx, getTokenByUserIDQuery, userID).Scan(
		&token.ID,
		&token.UserID,
		&token.Token,
		&token.TokenType,
		&token.IssuedAt,
		&token.ExpiredAt,
		&token.Duration,
		&token.CreatedAt,
		&token.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return token, nil
}
