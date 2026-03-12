package dto

import "time"

type UploadFileResponse struct {
	ID            uint      `json:"id"`
	UserID        uint      `json:"user_id"`
	FileName      string    `json:"file_name"`
	ObjectKey     string    `json:"object_key"`
	ContentType   string    `json:"content_type"`
	Size          int64     `json:"size"`
	WorkspacePath string    `json:"workspace_path"`
	CreatedAt     time.Time `json:"created_at"`
}
