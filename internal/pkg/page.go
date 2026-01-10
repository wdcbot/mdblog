package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Page struct {
	Title    string
	Slug     string
	Date     time.Time
	Content  string
	FilePath string
	Hidden   bool // 是否在导航栏隐藏
}

func ListPages() ([]Page, error) {
	basePath := filepath.Join("content", "page")
	files, err := os.ReadDir(basePath)
	if err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(basePath, 0755)
			return []Page{}, nil
		}
		return nil, err
	}

	var pages []Page
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".md" {
			filePath := filepath.Join(basePath, f.Name())
			post, err := ParseMarkdownFile(filePath)
			if err != nil { continue }
			
			// 读取 hidden 状态
			hidden := false
			content, _ := os.ReadFile(filePath)
			if strings.Contains(string(content), "hidden: true") {
				hidden = true
			}
			
			pages = append(pages, Page{
				Title:    post.Title,
				Slug:     post.Slug,
				Date:     post.Date,
				FilePath: filePath,
				Hidden:   hidden,
			})
		}
	}
	return pages, nil
}

// ListVisiblePages 获取导航栏显示的页面
func ListVisiblePages() ([]Page, error) {
	pages, err := ListPages()
	if err != nil {
		return nil, err
	}
	
	var visible []Page
	for _, p := range pages {
		if !p.Hidden {
			visible = append(visible, p)
		}
	}
	return visible, nil
}

func CreatePage(title string) (string, error) {
	slug := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
	path := filepath.Join("content", "page", slug+".md")
	
	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("page already exists")
	}

	content := "---\n" + 
	           "title: \"" + title + "\"\n" + 
	           "date: " + time.Now().Format("2006-01-02") + "\n" + 
	           "---\n\n" + 
	           "New Page Content"
	
	err := os.WriteFile(path, []byte(content), 0644)
	return path, err
}

func DeletePage(slug string) error {
	return os.Remove(filepath.Join("content", "page", slug+".md"))
}