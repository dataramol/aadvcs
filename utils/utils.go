package utils

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/dataramol/aadvcs/models"
)

const (
	Separator   = "|"
	defaultTime = "0001-01-01 00:00:00 +0000 UTC"
)

func CheckPathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CreateDirectories(dirs ...string) error {
	for _, dir := range dirs {
		if err := CreateDirectory(dir); err != nil {
			return err
		}
	}
	return nil
}

func CreateDirectory(dirName string) error {
	return os.MkdirAll(dirName, os.ModePerm)
}

func CreateFile(name string) error {
	_, err := CreateOrOpenFileRWMode(name)
	return err
}

func CreateOrOpenFileRWMode(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_CREATE|os.O_RDWR, os.ModePerm)
}

func CreateOrOpenFileAppendMode(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
}

func ClearFileContent(filePtr *os.File) error {
	err := filePtr.Truncate(0)
	filePtr.Seek(0, io.SeekStart)
	return err
}

func ExtractFileMetadataFromLine(lineStr string) models.FileMetaData {
	structure := strings.Split(lineStr, Separator)
	return models.FileMetaData{
		Path:             structure[0],
		ModificationTime: structure[1],
		Status:           models.FileStatus(structure[2]),
	}
}

func GetNumberOfChildrenDir(path string) (int, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return 0, err
	}
	return len(files), nil
}

func CreateNestedFile(p string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(p), os.ModePerm); err != nil {
		return nil, err
	}
	return os.Create(p)
}
