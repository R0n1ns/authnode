package domain

import "time"

// Role represents a user role in the system
type Role struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// DefaultRoles defines the roles available in the system
var DefaultRoles = struct {
	Admin string
	User  string
}{
	Admin: "admin",
	User:  "user",
}

// UserRole represents the association between a user and a role
type UserRole struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"userId" db:"user_id"`
	RoleID    int64     `json:"roleId" db:"role_id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}
