package file

import "time"

type File struct {
	ID            uint
	UserID        uint
	FileName      string
	ObjectKey     string
	ContentType   string
	Size          int64
	WorkspacePath string
	CreatedAt     time.Time
}
