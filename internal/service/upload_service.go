package service

import (
	"context"
	"io"
	"mime/multipart"
	"os"
	"path"

	"github.com/remiehneppo/material-management/utils"
)

type UploadService interface {
	UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, dir string, fileName string) (string, error)
}

type uploadService struct {
	baseDir string
}

func NewUploadService(baseDir string) UploadService {
	if baseDir == "" {
		baseDir = "uploads/"
	}
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
			panic("Failed to create base directory: " + err.Error())
		}
	}
	return &uploadService{
		baseDir: baseDir,
	}
}

func (s *uploadService) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, dir, fileName string) (string, error) {

	// check if dir exists, if not create it
	if _, err := os.Stat(path.Join(s.baseDir, dir)); os.IsNotExist(err) {
		if err := os.MkdirAll(path.Join(s.baseDir, dir), os.ModePerm); err != nil {
			return "", err
		}
	}
	ext := utils.GetFileExtension(fileHeader.Filename)
	if ext == "" {
		return "", os.ErrInvalid
	}
	filePath := path.Join(s.baseDir, dir, fileName+"."+ext)
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return "", err
	}

	return filePath, nil
}
