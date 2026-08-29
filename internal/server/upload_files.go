package server

import (
	"errors"
	"os"
)

func uploadPartName(uploadID string) (string, error) {
	if len(uploadID) != 32 {
		return "", errors.New("invalid upload id")
	}
	for _, c := range uploadID {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return "", errors.New("invalid upload id")
		}
	}
	return uploadID + ".part", nil
}

func writeUploadPart(uploadDir, uploadID string, data []byte) error {
	name, err := uploadPartName(uploadID)
	if err != nil {
		return err
	}
	root, err := os.OpenRoot(uploadDir)
	if err != nil {
		return err
	}
	defer root.Close()
	return root.WriteFile(name, data, 0o600)
}

func openUploadPart(uploadDir, uploadID string, flag int) (*os.File, error) {
	name, err := uploadPartName(uploadID)
	if err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(uploadDir)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	return root.OpenFile(name, flag, 0o600)
}

func removeUploadPart(uploadDir, uploadID string) error {
	name, err := uploadPartName(uploadID)
	if err != nil {
		return err
	}
	root, err := os.OpenRoot(uploadDir)
	if err != nil {
		return err
	}
	defer root.Close()
	err = root.Remove(name)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
