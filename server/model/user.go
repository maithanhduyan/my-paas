package model

import "time"

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"` // admin | member | viewer
	CreatedAt    time.Time `json:"created_at"`
}

type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthStatus struct {
	Authenticated bool  `json:"authenticated"`
	SetupRequired bool  `json:"setup_required"`
	User          *User `json:"user,omitempty"`
}
