package models

import "time"

type User struct {
	ID        int       `json:"id" db:"id"`
	Login     string    `json:"login" db:"login"`
	Password  string    `json:"password" db:"password"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type UserRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}
