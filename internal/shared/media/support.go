package media

import (
	"os"
	"path/filepath"
)

func hasValidExt(ext string) bool {
	for _, fileExt := range allowedExtensions {
		if fileExt == ext {
			return true
		}
	}

	return false
}

func hasInvalidExt(ext string) bool {
	return !hasValidExt(ext)
}

func GetStorageDir() string {
	dir, err := os.Getwd()
	folder := StorageDir

	if err != nil {
		// Handle the error appropriately.
		// Path defaults to a relative path if getting WD fails
		return "./" + folder
	}

	// Resolve the path at runtime.
	return filepath.Join(dir, StorageDir)
}

func GetMediaDir() string {
	return GetStorageDir() + "/" + Dir
}

func GetUsersImagesDir() string {
	return GetMediaDir() + "/" + UsersDir
}

func GetPostsImagesDir() string {
	return GetMediaDir() + "/" + PostsDir
}
