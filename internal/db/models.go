package db

import (
	"time"
)

type Project struct {
	ID        string    `gorm:"primaryKey"` // nanoID
	Name      string    // purely for convenience when listing all projects
	Password  string    // hashed version
	UserID    string    // owner of project
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
