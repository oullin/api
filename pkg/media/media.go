package media

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"os"
	"path/filepath"
	"strings"
)

const Dir = "media"
const UsersDir = "users"
const PostsDir = "posts"
const StorageDir = "storage"

var maxFileSize = int64(50 * 1024 * 1024) // 50 MB in bytes
var allowedExtensions = []string{".jpg", ".jpeg", ".png"}

type Media struct {
	file         []byte
	baseFileName string
	ext          string
	name         string
	path         string
	uniqueId     string
	size         int64
}

func NewMedia(uniqueId string, file []byte, baseFileName string) (*Media, error) {
	if len(file) <= 0 {
		return nil, errors.New("the given file is empty")
	}

	size := int64(len(file))
	ext := strings.ToLower(filepath.Ext(baseFileName))
	filename := uniqueId + "-" + uuid.New().String() + ext

	if size > maxFileSize {
		return nil, errors.New("the given file is too big")
	}

	if hasInvalidExt(ext) {
		return nil, errors.New("the given file type is invalid")
	}

	media := Media{
		uniqueId:     uniqueId,
		file:         file,
		baseFileName: baseFileName,
		ext:          ext,
		name:         filename,
		size:         size,
		path:         filepath.Join(GetUsersImagesDir(), filename),
	}

	return &media, nil
}

func (m *Media) Upload(directory string) error {
	prefix := m.uniqueId
	//value := m.name

	//if strings.HasPrefix(value, prefix) {
	if err := m.RemovePrefixedFiles(directory, prefix); err != nil {
		return errors.New("there was an error removing the file: " + err.Error())
	}
	//}

	err := os.WriteFile(m.path, m.file, 0644)

	if err != nil {
		return err
	}

	return nil
}

func (m *Media) RemovePrefixedFiles(dirPath, prefix string) error {
	entries, err := os.ReadDir(dirPath)

	if err != nil {
		return fmt.Errorf("error reading the directory '%s': %w", dirPath, err)
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), prefix) {
			fullPath := filepath.Join(dirPath, entry.Name())
			err := os.RemoveAll(fullPath)
			if err != nil {
				return fmt.Errorf("error removing file '%s' in directory '%s': %w", fullPath, dirPath, err)
			}
		}
	}

	return nil
}

func (m *Media) GetFileName() string {
	return m.name
}
func (m *Media) GetFilePath(prefix string) string {
	filePath := m.path

	dir := filepath.Dir(filePath)
	ext := filepath.Ext(filePath)
	base := filepath.Base(filePath)
	name := strings.TrimSuffix(base, ext)
	newName := fmt.Sprintf("%s-%s%s", prefix, name, ext)

	return filepath.Join(dir, newName)
}

func (m *Media) GetHeaderName() string {
	return m.baseFileName
}

func (m *Media) GetExtension() string {
	return m.ext
}
