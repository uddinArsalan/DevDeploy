package domain

import "time"

type CreateUser struct {
	Name     string
	Email    string
	Password string
}

type LoginUser struct {
	Email    string
	Password string
}

type UserRoles string

var (
	UserRole  UserRoles = "user"
	Moderator UserRoles = "moderator"
	Admin     UserRoles = "admin"
)

type User struct {
	ID           int64
	Name         string
	Email        string
	Role         UserRoles
	PasswordHash string
	CreatedAt    time.Time
}
