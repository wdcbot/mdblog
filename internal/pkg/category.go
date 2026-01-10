package pkg

import (
	"errors"
	"os"
	"path/filepath"
)

type CategoryInfo struct {
	Name      string
	PostCount int
}

func ListCategories() ([]CategoryInfo, error) {
	basePath := filepath.Join("content", "blog")
	entries, err := os.ReadDir(basePath)
	if err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(basePath, 0755)
			return []CategoryInfo{}, nil
		}
		return nil, err
	}

	var categories []CategoryInfo
	for _, entry := range entries {
		if entry.IsDir() {
			count := 0
			files, _ := os.ReadDir(filepath.Join(basePath, entry.Name()))
			for _, f := range files {
				if !f.IsDir() && filepath.Ext(f.Name()) == ".md" {
					count++
				}
			}
			categories = append(categories, CategoryInfo{
				Name:      entry.Name(),
				PostCount: count,
			})
		}
	}
	return categories, nil
}

func CreateCategory(name string) error {
	path := filepath.Join("content", "blog", name)
	return os.MkdirAll(path, 0755)
}

func RenameCategory(oldName, newName string) error {
	oldPath := filepath.Join("content", "blog", oldName)
	newPath := filepath.Join("content", "blog", newName)
	return os.Rename(oldPath, newPath)
}

func DeleteCategory(name string) error {
	path := filepath.Join("content", "blog", name)
	files, err := os.ReadDir(path)
	if err != nil { return err }
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".md" {
			return errors.New("category is not empty")
		}
	}
	return os.RemoveAll(path)
}


// MovePostToCategory 移动文章到其他分类
func MovePostToCategory(oldPath, newCategory string) (string, error) {
	// 获取文件名
	filename := filepath.Base(oldPath)
	
	// 构建新路径
	newPath := filepath.Join("content", "blog", newCategory, filename)
	
	// 确保目标分类存在
	if err := os.MkdirAll(filepath.Dir(newPath), 0755); err != nil {
		return "", err
	}
	
	// 移动文件
	if err := os.Rename(oldPath, newPath); err != nil {
		return "", err
	}
	
	return newPath, nil
}
