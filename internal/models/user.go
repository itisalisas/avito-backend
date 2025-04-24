package models

import "github.com/google/uuid"

type Role string

const (
	Moderator Role = "moderator"
	Employee  Role = "employee"
)

type User struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	Role     Role      `json:"role"`
	Password string
}
