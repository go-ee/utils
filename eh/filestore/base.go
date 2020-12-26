package filestore

import (
	"context"
	"os"
)

const DEFAULT_FOLDER_PERM os.FileMode = 0777
const DEFAULT_FILE_PERM os.FileMode = 0644

type Base struct {
	folder            string
	defaultFolderPerm os.FileMode
	defaultFilePerm   os.FileMode
}

func NewBase(folder string) *Base {
	return &Base{
		folder:            folder,
		defaultFolderPerm: DEFAULT_FOLDER_PERM,
		defaultFilePerm:   DEFAULT_FILE_PERM,
	}
}

func (s *Base) Close(ctx context.Context) {
}
