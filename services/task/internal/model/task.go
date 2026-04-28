package model

import (
	"time"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID        string     `json:"id"         gorm:"primaryKey;type:uuid"`
	UserID    string     `json:"user_id"    gorm:"type:uuid;not null;index"`
	Prompt    string     `json:"prompt"     gorm:"type:text;not null"`
	Status    TaskStatus `json:"status"     gorm:"type:varchar(20);default:'pending'"`
	Result    string     `json:"result"     gorm:"type:text"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
