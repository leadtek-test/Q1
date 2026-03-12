package adapters

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/leadtek-test/q1/common/consts"
	"github.com/leadtek-test/q1/common/handler/errors"
)

type WorkspaceRepositoryLocal struct {
	root string
}

func NewWorkspaceRepositoryLocal(root string) *WorkspaceRepositoryLocal {
	return &WorkspaceRepositoryLocal{root: root}
}

func (w WorkspaceRepositoryLocal) EnsureUserDir(userID uint) (string, error) {
	userDir := filepath.Join(w.root, strconv.FormatUint(uint64(userID), 10))
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		return "", errors.NewWithError(consts.ErrnoFileWorkspaceSaveFail, err)
	}
	return userDir, nil
}

func (w WorkspaceRepositoryLocal) Save(userID uint, fileName string, data []byte) (string, error) {
	userDir, err := w.EnsureUserDir(userID)
	if err != nil {
		return "", err
	}

	path := filepath.Join(userDir, filepath.Base(fileName))
	if err = os.WriteFile(path, data, 0o644); err != nil {
		return "", errors.NewWithError(consts.ErrnoFileWorkspaceSaveFail, err)
	}
	return path, nil
}
