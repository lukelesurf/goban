package model

import "time"

type Attachment struct {
	ID          string    `json:"id"`
	TaskID      string    `json:"task_id"`
	Filename    string    `json:"filename"`
	StoragePath string    `json:"-"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	CreatedAt   time.Time `json:"created_at"`
}
