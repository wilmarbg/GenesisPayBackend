package auth

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID	uuid.UUID	`gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Email	string		`gorm:"column:email;uniqueIndex;not null" json:"email"`

	PasswordHash	string		`gorm:"column:password_hash;not null" json:"-"`
	Role		string		`gorm:"column:role;not null;default:'CLIENTE'" json:"role"`
	IsActive	bool		`gorm:"column:is_active;not null;default:true" json:"is_active"`
	CreatedAt	time.Time	`gorm:"column:created_at" json:"created_at"`
	UpdatedAt	time.Time	`gorm:"column:updated_at" json:"updated_at"`
}

func (User) TableName() string {
	return `"aut".users`
}

type RegisterRequest struct {
	Email		string	`json:"email" binding:"required"`
	Password	string	`json:"password" binding:"required"`
	Role		string	`json:"role" binding:"required"`
}

type LoginRequest struct {
	Email		string	`json:"email" binding:"required"`
	Password	string	`json:"password" binding:"required"`
}

type AuthResponse struct {
	User	User	`json:"user"`
	Token	string	`json:"token"`
}
