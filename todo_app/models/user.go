package models

import "time"

type User struct {
	Id          int       `json:"id"`
	Email       string    `json:"email" validate:"required,email"`
	Username    string    `json:"username" validate:"required"`
	Password    string    `json:"password" validate:"required"`
	Facebook_id string    `json:"facebook_id"`
	Created_at  time.Time `json:"created_at"`
	Updated_at  time.Time `json:"updated_at"`
}

type UserLogin struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}
