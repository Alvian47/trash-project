package models

import "time"

type Task struct {
	Id         int       `json:"id"`
	Title      string    `json:"title" validate:"required"`
	Content    string    `json:"content" validate:"required"`
	Status     bool      `json:"status" validate:"required"`
	Deadline   time.Time `json:"deadline" validate:"required"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
}
