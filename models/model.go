package models

import (
	"database/sql"
	"gorm.io/gorm"
)

type Model struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	CreatedAt sql.NullTime   `json:"created_at"`
	UpdatedAt sql.NullTime   `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}
