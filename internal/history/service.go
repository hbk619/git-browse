package history

import (
	"encoding/json"
	"github.com/hbk619/git-browse/internal/filesystem"
	"os"
	"path"
)

type (
	Storage interface {
		Load() (History, error)
		Save(h History) error
	}
	Service struct {
		configPath string
		fs         filesystem.FS
	}

	PR struct {
		CommentCount int
	}

	History struct {
		Prs map[int]PR
	}
)

func NewHistoryService(basePath string, fs filesystem.FS) (*Service, error) {
	configPath := path.Join(basePath, ".config")
	err := fs.MkdirAll(configPath, os.ModeDir)
	if err != nil {
		return nil, err
	}
	return &Service{
		configPath: path.Join(configPath, "git-browse-history.json"),
		fs:         fs,
	}, nil
}

func (service *Service) Load() (History, error) {
	file, err := service.fs.ReadFile(service.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return History{
				Prs: make(map[int]PR),
			}, nil
		}
		return History{}, err
	}

	var history History
	err = json.Unmarshal(file, &history)
	if err != nil {
		return History{}, err
	}
	return history, nil
}

func (service *Service) Save(history History) error {
	marshalled, err := json.Marshal(history)
	if err != nil {
		return err
	}

	return service.fs.SaveFile(service.configPath, marshalled)
}
