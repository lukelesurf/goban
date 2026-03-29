package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lukelesurf/goban/internal/db"
	"github.com/lukelesurf/goban/internal/model"
)

var allowedExtensions = map[string]bool{
	".txt": true, ".md": true, ".pdf": true,
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true, ".svg": true,
	".zip": true, ".json": true, ".yaml": true, ".yml": true,
	".csv": true, ".log": true, ".xml": true, ".html": true,
}

type AttachmentService struct {
	db         *db.DB
	storageDir string
}

func NewAttachmentService(db *db.DB, storageDir string) *AttachmentService {
	os.MkdirAll(storageDir, 0755)
	return &AttachmentService{db: db, storageDir: storageDir}
}

func (s *AttachmentService) Upload(ctx context.Context, taskID, filename string, r io.Reader) (*model.Attachment, error) {
	// Validate task exists
	var exists int
	if err := s.db.QueryRowContext(ctx, `SELECT 1 FROM tasks WHERE id = ?`, taskID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	// Validate extension
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedExtensions[ext] {
		return nil, fmt.Errorf("file type %q not allowed", ext)
	}

	// Sanitise filename
	safeName := filepath.Base(filename)

	id := uuid.NewString()
	relPath := filepath.Join(taskID, id+"-"+safeName)
	fullPath := filepath.Join(s.storageDir, relPath)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, fmt.Errorf("create attachment dir: %w", err)
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	// Buffer first 512 bytes for content type detection
	buf := make([]byte, 512)
	n, err := io.ReadFull(r, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		os.Remove(fullPath)
		return nil, fmt.Errorf("read file: %w", err)
	}
	buf = buf[:n]
	contentType := http.DetectContentType(buf)

	written, err := f.Write(buf)
	if err != nil {
		os.Remove(fullPath)
		return nil, fmt.Errorf("write file: %w", err)
	}

	copied, err := io.Copy(f, r)
	if err != nil {
		os.Remove(fullPath)
		return nil, fmt.Errorf("write file: %w", err)
	}
	size := int64(written) + copied

	now := time.Now().UTC()
	att := model.Attachment{
		ID:          id,
		TaskID:      taskID,
		Filename:    safeName,
		StoragePath: relPath,
		Size:        size,
		ContentType: contentType,
		CreatedAt:   now,
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO attachments (id, task_id, filename, storage_path, size, content_type, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		att.ID, att.TaskID, att.Filename, att.StoragePath, att.Size, att.ContentType, att.CreatedAt,
	)
	if err != nil {
		os.Remove(fullPath)
		return nil, fmt.Errorf("insert attachment: %w", err)
	}

	return &att, nil
}

func (s *AttachmentService) UploadFromString(ctx context.Context, taskID, filename, content string) (*model.Attachment, error) {
	return s.Upload(ctx, taskID, filename, strings.NewReader(content))
}

func (s *AttachmentService) List(ctx context.Context, taskID string) ([]model.Attachment, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, task_id, filename, storage_path, size, content_type, created_at
		 FROM attachments WHERE task_id = ? ORDER BY created_at ASC`, taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("list attachments: %w", err)
	}
	defer rows.Close()

	var attachments []model.Attachment
	for rows.Next() {
		var a model.Attachment
		if err := rows.Scan(&a.ID, &a.TaskID, &a.Filename, &a.StoragePath, &a.Size, &a.ContentType, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan attachment: %w", err)
		}
		attachments = append(attachments, a)
	}
	return attachments, rows.Err()
}

func (s *AttachmentService) Get(ctx context.Context, id string) (*model.Attachment, error) {
	var a model.Attachment
	err := s.db.QueryRowContext(ctx,
		`SELECT id, task_id, filename, storage_path, size, content_type, created_at
		 FROM attachments WHERE id = ?`, id,
	).Scan(&a.ID, &a.TaskID, &a.Filename, &a.StoragePath, &a.Size, &a.ContentType, &a.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get attachment %s: %w", id, err)
	}
	return &a, nil
}

func (s *AttachmentService) FilePath(a *model.Attachment) string {
	return filepath.Join(s.storageDir, a.StoragePath)
}

func (s *AttachmentService) Delete(ctx context.Context, id string) error {
	a, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `DELETE FROM attachments WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete attachment %s: %w", id, err)
	}

	os.Remove(filepath.Join(s.storageDir, a.StoragePath))
	os.Remove(filepath.Join(s.storageDir, a.TaskID)) // remove dir if empty
	return nil
}

func (s *AttachmentService) DeleteAllForTask(ctx context.Context, taskID string) error {
	os.RemoveAll(filepath.Join(s.storageDir, taskID))
	return nil
}
