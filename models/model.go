package models

import (
	"database/sql"
	"gorm.io/gorm"
)

type Model struct {
	ID        int64          `gorm:"primaryKey" json:"id" form:"id"`
	CreatedAt sql.NullTime   `json:"created_at" form:"created_at"`
	UpdatedAt sql.NullTime   `json:"updated_at" form:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" form:"deleted_at"`
}
