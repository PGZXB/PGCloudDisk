package models

const (
	FILE_TYPE = "FILE"
	DIR_TYPE  = "DIR"
)

type File struct {
	Model
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
	Location  string `json:"location"`
	LocalAddr string `json:"local_addr"`
	Type      string `json:"type"`
	UserID    int64  `json:"user_id"`
}