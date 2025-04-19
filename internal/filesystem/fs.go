package filesystem

import "os"

type FS interface {
	MkdirAll(path string, perm os.FileMode) error
	ReadFile(filePath string) ([]byte, error)
	SaveFile(filePath string, data []byte) error
}

type FileSystem struct{}

func NewFS() *FileSystem {
	return &FileSystem{}
}

func (fs *FileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (fs *FileSystem) ReadFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

func (fs *FileSystem) SaveFile(filePath string, data []byte) error {
	return os.WriteFile(filePath, data, 0644)
}
