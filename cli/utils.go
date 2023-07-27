package cli

import (
	"io"
	"os"
	"strings"

	"github.com/dataramol/aadvcs/models"
)

const (
	separator   = "|"
	defaultTime = "0001-01-01 00:00:00 +0000 UTC"
)

func checkPathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func createDirectories(dirs ...string) error {
	for _, dir := range dirs {
		if err := createDirectory(dir); err != nil {
			return err
		}
	}
	return nil
}

func createDirectory(dirName string) error {
	return os.MkdirAll(dirName, os.ModePerm)
}

func createFile(name string) error {
	_, err := createOrOpenFileRWMode(name)
	return err
}

func createOrOpenFileRWMode(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_CREATE|os.O_RDWR, os.ModePerm)
}

func CreateOrOpenFileAppendMode(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
}

func clearFileContent(filePtr *os.File) error {
	err := filePtr.Truncate(0)
	filePtr.Seek(0, io.SeekStart)
	return err
}

func extractFileMetadataFromLine(lineStr string) models.FileMetaData {
	structure := strings.Split(lineStr, separator)
	return models.FileMetaData{
		Path:             structure[0],
		ModificationTime: structure[1],
		Status:           models.FileStatus(structure[2]),
	}
}
