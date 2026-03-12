package command

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/leadtek-test/q1/common/consts"
	"github.com/leadtek-test/q1/common/decorator"
	"github.com/leadtek-test/q1/common/handler/errors"
	"github.com/leadtek-test/q1/container/domain/file"
	"github.com/sirupsen/logrus"
)

const defaultFileContentType = "application/octet-stream"

type UploadFile struct {
	UserID      uint   `json:"user_id"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	Payload     []byte `json:"-"`
}

type UploadFileResult struct {
	ID            uint
	UserID        uint
	FileName      string
	ObjectKey     string
	ContentType   string
	Size          int64
	WorkspacePath string
	CreatedAt     time.Time
}

type UploadFileHandler decorator.CommandHandler[UploadFile, *UploadFileResult]

type uploadFileHandler struct {
	repo        file.Repository
	objectStore file.ObjectStorage
	workspace   file.Workspace
	maxFileSize int64
}

func NewUploadFileHandler(
	repo file.Repository,
	objectStore file.ObjectStorage,
	workspace file.Workspace,
	maxFileSize int64,
	logger *logrus.Logger,
) UploadFileHandler {
	if repo == nil {
		panic("upload file's repository is nil")
	}
	if objectStore == nil {
		panic("upload file's object storage is nil")
	}
	if workspace == nil {
		panic("upload file's workspace is nil")
	}
	if maxFileSize <= 0 {
		panic("upload file's max file size must be greater than 0")
	}

	return decorator.ApplyCommandDecorators(
		uploadFileHandler{
			repo:        repo,
			objectStore: objectStore,
			workspace:   workspace,
			maxFileSize: maxFileSize,
		},
		logger,
	)
}

func (h uploadFileHandler) Handle(ctx context.Context, cmd UploadFile) (*UploadFileResult, error) {
	fileName := strings.TrimSpace(cmd.FileName)
	contentType := strings.TrimSpace(cmd.ContentType)
	if contentType == "" {
		contentType = defaultFileContentType
	}

	if err := h.validate(cmd.UserID, fileName, int64(len(cmd.Payload))); err != nil {
		return nil, err
	}

	cleanName := filepath.Base(fileName)
	fileID := newFileID()
	storeName := fileID + "_" + cleanName
	objectKey := fmt.Sprintf("users/%d/%s", cmd.UserID, storeName)
	size := int64(len(cmd.Payload))

	if err := h.objectStore.Upload(ctx, objectKey, bytes.NewReader(cmd.Payload), size, contentType); err != nil {
		return nil, errors.NewWithError(consts.ErrnoFileUploadFailed, err)
	}

	workspacePath, err := h.workspace.Save(cmd.UserID, storeName, cmd.Payload)
	if err != nil {
		return nil, errors.NewWithError(consts.ErrnoFileWorkspaceSaveFail, err)
	}

	data := &file.File{
		UserID:        cmd.UserID,
		FileName:      cleanName,
		ObjectKey:     objectKey,
		ContentType:   contentType,
		Size:          size,
		WorkspacePath: workspacePath,
	}

	if err = h.repo.Create(ctx, data); err != nil {
		return nil, err
	}

	return &UploadFileResult{
		ID:            data.ID,
		UserID:        data.UserID,
		FileName:      data.FileName,
		ObjectKey:     data.ObjectKey,
		ContentType:   data.ContentType,
		Size:          data.Size,
		WorkspacePath: data.WorkspacePath,
		CreatedAt:     data.CreatedAt,
	}, nil
}

func (h uploadFileHandler) validate(userID uint, fileName string, size int64) error {
	if userID == 0 {
		return errors.New(consts.ErrnoAuthInvalidToken)
	}
	if fileName == "" {
		return errors.New(consts.ErrnoFileNameRequired)
	}
	if size > h.maxFileSize {
		return errors.New(consts.ErrnoFileSizeExceeded)
	}
	return nil
}

func newFileID() string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}
