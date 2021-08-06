package models

import (
	"time"
)

const (
	FileType = "FILE"
	DirType  = "DIR"
)

type File struct {
	Model
	Filename  string `json:"filename" form:"filename"`
	Size      int64  `json:"size" form:"size"`
	Location  string `json:"location" form:"location"`
	LocalAddr string `json:"local_addr" form:"local_addr"`
	Type      string `json:"type" form:"type"`
	UserID    int64  `json:"user_id" form:"user_id"`
}

type FileInfoCanBeUpdated struct {
	Filename string `json:"filename" form:"filename"`
}

type FileInfoCanBePublished struct {
	ID        int64     `json:"id,omitempty" form:"id"`
	CreatedAt time.Time `json:"created_at,omitempty" form:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty" form:"updated_at"`
	DeletedAt time.Time `json:"deleted_at,omitempty" form:"deleted_at"`
	Filename  string    `json:"filename,omitempty" form:"filename"`
	Size      int64     `json:"size,omitempty" form:"size"`
	Location  string    `json:"location,omitempty" form:"location"`
	Type      string    `json:"type,omitempty" form:"type"`
}
