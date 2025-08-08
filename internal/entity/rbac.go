package entity

import (
	"database/sql/driver"

	"github.com/elangreza/content-management-system/internal/constanta"
)

type UserRole struct {
	name        string
	val         int64
	permissions []constanta.UserPermission
}

func NewUserRole(name string, permissions ...constanta.UserPermission) UserRole {
	var val int64
	for _, permission := range permissions {
		val |= int64(permission)
	}
	return UserRole{
		name:        name,
		val:         val,
		permissions: permissions,
	}
}

func (r UserRole) HasPermission(permission constanta.UserPermission) bool {
	return r.val&int64(permission) != 0
}

// Scan implements the sql.Scanner interface for UserRole. from db with int value in database.
func (r *UserRole) Scan(src interface{}) error {
	if val, ok := src.(int64); ok {
		r.val = val
		return nil
	}
	return nil
}

// value implements the driver.Valuer interface for UserRole. to db with int value in database.
func (r UserRole) Value() (driver.Value, error) {
	return r.val, nil
}

func (r UserRole) GetValue() int64 {
	return r.val
}

func (r UserRole) GetPermissions() []constanta.UserPermission {
	if len(r.permissions) > 0 {
		return r.permissions
	}

	// If permissions are not set, derive them from val
	permissions := make([]constanta.UserPermission, 0)
	for _, p := range permissions {
		if r.val&int64(p) != 0 {
			permissions = append(permissions, p)
		}
	}
	return permissions
}

func (r UserRole) GetName() string {
	return r.name
}
