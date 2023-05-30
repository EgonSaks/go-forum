package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"forum/logger"
)

func (app *application) UploadImage(image multipart.File, extension string) (string, error) {
	destinationFolder := "ui/static/img/uploads/post"

	// Create the destination folder if it does not exist
	if err := os.MkdirAll(destinationFolder, os.ModePerm); err != nil {
		logger.ErrorLogger.Printf("Error creating destination folder: %v\n", err)
		return "", err
	}

	// Calculate the hash of the file name
	hash := sha256.New()
	if _, err := io.Copy(hash, image); err != nil {
		logger.ErrorLogger.Printf("Error calculating file hash: %v\n", err)
		return "", err
	}
	imageHash := hex.EncodeToString(hash.Sum(nil))

	// Create a new file in the destination folder and write the uploaded file to it
	fileName := imageHash + extension
	fullPath := filepath.Join(destinationFolder, fileName)

	// Check if file already exists
	// if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
	// 	return "", errors.New("image already exists")
	// }

	newFile, err := os.Create(fullPath)
	if err != nil {
		logger.ErrorLogger.Printf("Error creating new file: %v\n", err)
		return "", err
	}
	defer newFile.Close()

	// Set the file pointer back to the beginning of the file
	if _, err := image.Seek(0, 0); err != nil {
		logger.ErrorLogger.Printf("Error setting file pointer to the beginning: %v\n", err)
		return "", err
	}

	if _, err := io.Copy(newFile, image); err != nil {
		logger.ErrorLogger.Printf("Error copying file: %v\n", err)
		return "", err
	}

	logger.InfoLogger.Printf("Successfully uploaded file to %v\n", fullPath)

	// Return the file path
	return strings.Replace(fullPath, "ui/", "", 1), nil
}
