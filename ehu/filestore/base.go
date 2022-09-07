package filestore

import (
	"context"
	"os"
)

const DefaultFolderPerm os.FileMode = 0777
const DefaultFilePerm os.FileMode = 0644

type Base struct {
	folder            string
	defaultFolderPerm os.FileMode
	defaultFilePerm   os.FileMode
}

func NewBase(folder string) *Base {
	return &Base{
		folder:            folder,
		defaultFolderPerm: DefaultFolderPerm,
		defaultFilePerm:   DefaultFilePerm,
	}
}

func (s *Base) Close(ctx context.Context) {
}
