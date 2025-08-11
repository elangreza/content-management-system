package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/elangreza/content-management-system/internal/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserRepo_CreateUser(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(createUserQuery)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "fail",
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(createUserQuery)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("insert error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, _ := sqlmock.New()
			defer db.Close()
			repo := NewUserRepo(db)
			if tt.prepare != nil {
				tt.prepare(mock)
			}
			user := entity.User{}
			err := repo.CreateUser(context.Background(), user)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserRepo_GetUserByEmail(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			prepare: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
					AddRow(uuid.New(), "test", "test@mail.com", []byte("pass"), "admin", time.Now(), sql.NullTime{})
				mock.ExpectQuery(regexp.QuoteMeta(getUserByEmailQuery)).
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "fail",
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(getUserByEmailQuery)).
					WithArgs(sqlmock.AnyArg()).
					WillReturnError(errors.New("query error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, _ := sqlmock.New()
			defer db.Close()
			repo := NewUserRepo(db)
			if tt.prepare != nil {
				tt.prepare(mock)
			}
			_, err := repo.GetUserByEmail(context.Background(), "test@mail.com")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserRepo_GetUserByID(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			prepare: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
					AddRow(uuid.New(), "test", "test@mail.com", []byte("pass"), "admin", time.Now(), sql.NullTime{})
				mock.ExpectQuery(regexp.QuoteMeta(getUserByIDQuery)).
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "fail",
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(getUserByIDQuery)).
					WithArgs(sqlmock.AnyArg()).
					WillReturnError(errors.New("query error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, _ := sqlmock.New()
			defer db.Close()
			repo := NewUserRepo(db)
			if tt.prepare != nil {
				tt.prepare(mock)
			}
			_, err := repo.GetUserByID(context.Background(), uuid.New())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserRepo_GetUserRoleByUserID(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			prepare: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"role"}).AddRow("admin")
				mock.ExpectQuery(regexp.QuoteMeta(getUserRoleByUserIDQuery)).
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "fail",
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(getUserRoleByUserIDQuery)).
					WithArgs(sqlmock.AnyArg()).
					WillReturnError(errors.New("query error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, _ := sqlmock.New()
			defer db.Close()
			repo := NewUserRepo(db)
			if tt.prepare != nil {
				tt.prepare(mock)
			}
			_, err := repo.GetUserRoleByUserID(context.Background(), uuid.New())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
