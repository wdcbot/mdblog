package pkg

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type Post struct {
	Title       string
	Slug        string
	Date        time.Time
	Tags        []string
	Category    string
	Content     string
	Summary     string
	FilePath    string
	ReadingTime int       // 阅读时间（分钟）
	WordCount   int       // 字数统计
	TOC         []TOCItem // 文章目录
	Draft       bool      // 是否为草稿
	Pinned      bool      // 是否置顶
}

// TOCItem 目录项
type TOCItem struct {
	Level int
	ID    string
	Title string
}

var mdProcessor goldmark.Markdown

func InitMarkdown() {
	mdProcessor = goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
}

func ParseMarkdownFile(path string) (*Post, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	context := parser.NewContext()
	if err := mdProcessor.Convert(content, &buf, parser.WithContext(context)); err != nil {
		return nil, err
	}

	metaData := meta.Get(context)
	
	post := &Post{
		Content:  buf.String(),
		FilePath: path,
	}

	if title, ok := metaData["title"].(string); ok {
		post.Title = title
	} else {
		post.Title = filepath.Base(path)
	}

	if dateStr, ok := metaData["date"].(string); ok {
		t, _ := time.Parse("2006-01-02", dateStr)
		post.Date = t
	}

	if tags, ok := metaData["tags"].([]interface{}); ok {
		for _, t := range tags {
			if ts, ok := t.(string); ok {
				post.Tags = append(post.Tags, ts)
			}
		}
	}

	// 解析草稿状态
	if draft, ok := metaData["draft"].(bool); ok {
		post.Draft = draft
	}

	// 解析置顶状态
	if pinned, ok := metaData["pinned"].(bool); ok {
		post.Pinned = pinned
	}

	// Category Logic: Calculate relative path from "content/blog"
	// Example: content/blog/tech/go.md -> "tech"
	// Example: content/blog/life.md -> "uncategorized"
	
	// Normalize path separators
	cleanPath := filepath.Clean(path)
	// Create base path for comparison
	baseBlogPath := filepath.Join("content", "blog")
	
	relPath, err := filepath.Rel(baseBlogPath, cleanPath)
	if err == nil {
		dir := filepath.Dir(relPath)
		if dir == "." || dir == "" {
			post.Category = "uncategorized"
		} else {
			// Extract the first level directory as category
			// e.g. "tech/backend" -> "tech"
			parts := strings.Split(filepath.ToSlash(dir), "/")
			if len(parts) > 0 {
				post.Category = parts[0]
			}
		}
	} else {
		// Fallback
		post.Category = filepath.Base(filepath.Dir(cleanPath))
	}

	// Slug priority: Frontmatter > Filename
	if customSlug, ok := metaData["slug"].(string); ok && customSlug != "" {
		post.Slug = customSlug
	} else {
		post.Slug = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}

	// 打印加载日志，方便排查 404
	fmt.Printf("[DEBUG] Loaded Post: Slug=%s, Category=%s, Path=%s\n", post.Slug, post.Category, path)

	// 自动生成摘要
	post.Summary = generateSummary(post.Content, 120)

	// 计算阅读时间和字数
	post.WordCount, post.ReadingTime = calculateWordStats(post.Content)

	// 生成目录并添加标题 ID
	post.Content, post.TOC = generateTOC(post.Content)

	return post, nil
}

// generateSummary 从 HTML 内容中提取纯文本摘要
func generateSummary(htmlContent string, maxLen int) string {
	// 移除 HTML 标签
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(htmlContent, "")
	
	// 移除多余空白
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)
	
	// 按字符数截断（支持中文）
	if utf8.RuneCountInString(text) <= maxLen {
		return text
	}
	
	runes := []rune(text)
	return string(runes[:maxLen]) + "..."
}

func ReadPostFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func SavePostFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func CreatePostFile(category, title, slug string, draft bool) (string, error) {
	// Normalize slug: if empty, derive from title; keep only [a-z0-9-]
	if strings.TrimSpace(slug) == "" {
		slug = strings.ToLower(title)
		slug = strings.ReplaceAll(slug, " ", "-")
	}
	slug = strings.ToLower(slug)
	// remove non alphanumeric and dash
	re := regexp.MustCompile(`[^a-z0-9\-]+`)
	slug = re.ReplaceAllString(slug, "")
	// collapse multiple dashes
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		// fallback to timestamp-based slug for non-latin titles
		slug = time.Now().Format("20060102-150405")
	}
	
	fileName := slug + ".md"
	filePath := filepath.Join("content", "blog", category, fileName)

	// Check if already exists
	if _, err := os.Stat(filePath); err == nil {
		return "", errors.New("a post with this slug already exists in this category")
	}
	
	// 构建 frontmatter
	var content string
	if draft {
		content = fmt.Sprintf("---\ntitle: \"%s\"\ndate: %s\ndraft: true\ntags: []\n---\n\nStart writing here...", 
			title, time.Now().Format("2006-01-02"))
	} else {
		content = fmt.Sprintf("---\ntitle: \"%s\"\ndate: %s\ntags: []\n---\n\nStart writing here...", 
			title, time.Now().Format("2006-01-02"))
	}
	
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func DeletePostFile(path string) error {
	return os.Remove(path)
}


// calculateWordStats 计算字数和阅读时间
func calculateWordStats(htmlContent string) (wordCount int, readingTime int) {
	// 移除 HTML 标签
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(htmlContent, "")
	text = strings.TrimSpace(text)

	// 统计中文字符数
	chineseCount := 0
	for _, r := range text {
		if r >= 0x4e00 && r <= 0x9fff {
			chineseCount++
		}
	}

	// 统计英文单词数
	words := regexp.MustCompile(`[a-zA-Z]+`).FindAllString(text, -1)
	englishWords := len(words)

	// 总字数
	wordCount = chineseCount + englishWords

	// 计算阅读时间（中文 300 字/分钟，英文 200 词/分钟）
	minutes := float64(chineseCount)/300 + float64(englishWords)/200
	if minutes < 1 {
		readingTime = 1
	} else {
		readingTime = int(minutes + 0.5)
	}
	return
}


// generateTOC 从 HTML 中提取标题生成目录，并为标题添加 ID
func generateTOC(htmlContent string) (string, []TOCItem) {
	var toc []TOCItem
	
	// 匹配 h2, h3 标签
	re := regexp.MustCompile(`<(h[23])>([^<]+)</h[23]>`)
	
	counter := 0
	result := re.ReplaceAllStringFunc(htmlContent, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}
		
		tag := submatches[1]
		title := strings.TrimSpace(submatches[2])
		
		level := 2
		if tag == "h3" {
			level = 3
		}
		
		counter++
		id := fmt.Sprintf("heading-%d", counter)
		
		toc = append(toc, TOCItem{
			Level: level,
			ID:    id,
			Title: title,
		})
		
		return fmt.Sprintf(`<%s id="%s">%s</%s>`, tag, id, title, tag)
	})
	
	return result, toc
}

// RenderMarkdownPreview 渲染 Markdown 内容为 HTML（用于预览）
func RenderMarkdownPreview(content string) string {
	var buf bytes.Buffer
	context := parser.NewContext()
	if err := mdProcessor.Convert([]byte(content), &buf, parser.WithContext(context)); err != nil {
		return "<p>渲染失败</p>"
	}
	return buf.String()
}
