package entity

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       uuid.UUID `db:"id"`
	Email    string    `db:"email"`
	Name     string    `db:"name"`
	password []byte    `db:"password"`
	// default 0
	Permission int

	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt sql.NullTime `db:"updated_at"`
}

func NewUser(email, password, name string) (*User, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	pass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:       id,
		Email:    email,
		Name:     name,
		password: pass,
	}, nil
}

func (u *User) IsPasswordValid(reqPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.password), []byte(reqPassword))
	return err == nil
}

func (u *User) GetPassword() []byte {
	return u.password
}

func (u *User) SetPassword(password []byte) {
	u.password = password
}

// func (u *User) ValidPermission(reqPermission int) bool {
// 	return (u.Permission & reqPermission) > 0
// }

// var UserValPermission = UserPermission{
// 	Val:  1,
// 	Name: "user",
// }

// var AdminValPermission = UserPermission{
// 	Val:  2,
// 	Name: "admin",
// }

// var DefaultUserPermissions = []UserPermission{UserValPermission, AdminValPermission}

// func (u *User) LoadPermissions() {
// 	if len(u.Permissions) > 0 {
// 		return
// 	}

// 	for _, permission := range DefaultUserPermissions {
// 		if u.ValidPermission(permission.Val) && permission.Val <= u.Permission {
// 			u.Permissions = append(u.Permissions, permission)
// 		}
// 	}
// }
