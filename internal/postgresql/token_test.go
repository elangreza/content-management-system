package postgresql

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/elangreza/content-management-system/internal/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTokenRepo_CreateToken(t *testing.T) {
	type args struct {
		ctx   context.Context
		token entity.Token
	}

	tests := []struct {
		name    string
		args    args
		mock    func(sqlmock.Sqlmock, args)
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				token: entity.Token{
					ID:        uuid.New(),
					UserID:    uuid.New(),
					Token:     "token",
					TokenType: "access",
					IssuedAt:  time.Now(),
					ExpiredAt: time.Now().Add(time.Hour),
					Duration:  time.Hour.String(),
				},
			},
			mock: func(m sqlmock.Sqlmock, a args) {
				m.ExpectExec(regexp.QuoteMeta(createTokenQuery)).
					WithArgs(
						a.token.ID,
						a.token.UserID,
						a.token.Token,
						a.token.TokenType,
						a.token.IssuedAt,
						a.token.ExpiredAt,
						a.token.Duration,
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "fail",
			args: args{
				ctx:   context.Background(),
				token: entity.Token{},
			},
			mock: func(m sqlmock.Sqlmock, a args) {
				m.ExpectExec(regexp.QuoteMeta(createTokenQuery)).
					WithArgs(
						a.token.ID,
						a.token.UserID,
						a.token.Token,
						a.token.TokenType,
						a.token.IssuedAt,
						a.token.ExpiredAt,
						a.token.Duration,
					).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, _ := sqlmock.New()
			repo := NewTokenRepo(db)
			defer db.Close()
			tt.mock(mock, tt.args)
			err := repo.CreateToken(tt.args.ctx, tt.args.token)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTokenRepo_GetTokenByTokenID(t *testing.T) {
	type args struct {
		ctx     context.Context
		tokenID uuid.UUID
	}

	tests := []struct {
		name    string
		args    args
		mock    func(sqlmock.Sqlmock, args)
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:     context.Background(),
				tokenID: uuid.New(),
			},
			mock: func(m sqlmock.Sqlmock, a args) {
				rows := sqlmock.NewRows([]string{"id", "user_id", "token", "token_type", "issued_at", "expired_at", "duration"}).
					AddRow(a.tokenID, uuid.New(), "token", "access", time.Now(), time.Now().Add(time.Hour), time.Hour.String())
				m.ExpectQuery(regexp.QuoteMeta(getTokenByTokenIDQuery)).
					WithArgs(a.tokenID).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "fail",
			args: args{
				ctx:     context.Background(),
				tokenID: uuid.New(),
			},
			mock: func(m sqlmock.Sqlmock, a args) {
				m.ExpectQuery(regexp.QuoteMeta(getTokenByTokenIDQuery)).
					WithArgs(a.tokenID).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, _ := sqlmock.New()
			repo := NewTokenRepo(db)
			defer db.Close()
			tt.mock(mock, tt.args)
			_, err := repo.GetTokenByTokenID(tt.args.ctx, tt.args.tokenID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTokenRepo_GetTokenByUserID(t *testing.T) {
	type args struct {
		ctx    context.Context
		userID uuid.UUID
	}

	tests := []struct {
		name    string
		args    args
		mock    func(sqlmock.Sqlmock, args)
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
			},
			mock: func(m sqlmock.Sqlmock, a args) {
				rows := sqlmock.NewRows([]string{"id", "user_id", "token", "token_type", "issued_at", "expired_at", "duration"}).
					AddRow(uuid.New(), a.userID, "token", "access", time.Now(), time.Now().Add(time.Hour), time.Hour.String())
				m.ExpectQuery(regexp.QuoteMeta(getTokenByUserIDQuery)).
					WithArgs(a.userID).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "fail",
			args: args{
				ctx:    context.Background(),
				userID: uuid.New(),
			},
			mock: func(m sqlmock.Sqlmock, a args) {
				m.ExpectQuery(regexp.QuoteMeta(getTokenByUserIDQuery)).
					WithArgs(a.userID).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, _ := sqlmock.New()
			repo := NewTokenRepo(db)
			defer db.Close()
			tt.mock(mock, tt.args)
			_, err := repo.GetTokenByUserID(tt.args.ctx, tt.args.userID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
