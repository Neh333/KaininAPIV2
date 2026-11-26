package model

import (
	"time"
)

type User struct {
	UserID            string     `json:"id,omitempty" db:"id"`
	FirstName         string     `json:"firstName" db:"first_name"`
	LastName          string     `json:"lastName" db:"last_name"`
	Email             string     `json:"email" db:"email"`
	isAdmin           bool       `json:"isAdmin" db:"is_admin"`
	PhoneNumber       string     `json:"phoneNumber" db:"phone_number"`
	RoleID            string     `json:"roleId" db:"role_id"`
	SignUpMethod      string     `json:"signUpMethod" db:"sign_up_method"`
	PhoneVerified     bool       `json:"phoneVerified" db:"phone_verified"`
	LastLogin         *time.Time `json:"lastLogin,omitempty" db:"last_login"`
	SuperTokensUserID string     `json:"superTokensUserId" db:"super_tokens_user_id"`
	Password          string     `json:"-" db:"password"`
	CreatedAt         time.Time  `json:"-" db:"created_at"`
	UpdatedAt         time.Time  `json:"-" db:"updated_at"`
}
