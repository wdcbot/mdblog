package pkg

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/blevesearch/bleve/v2"
)

var (
	Posts     []*Post
	PostsMap  map[string]*Post
	Index     bleve.Index
	storeLock sync.RWMutex

	// ContentCache 缓存已解析的 HTML 内容
	// Key: FilePath, Value: string (HTML)
	contentCache sync.Map
)

func InitStore() {
	PostsMap = make(map[string]*Post)
	LoadAllPosts()
	InitSearchIndex()
}

// LoadAllPosts 将 Markdown 元数据载入内存，按时间倒序排列
func LoadAllPosts() {
	storeLock.Lock()
	defer storeLock.Unlock()

	Posts = nil
	PostsMap = make(map[string]*Post)

	basePath := filepath.Join("content", "blog")
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		post, err := ParseMarkdownFile(path)
		if err != nil {
			return nil
		}
		
		// 草稿文章只加入 PostsMap（后台可见），不加入 Posts（前台不可见）
		key := strings.ToLower(post.Category + "/" + post.Slug)
		PostsMap[key] = post
		
		if !post.Draft {
			Posts = append(Posts, post)
		}
		return nil
	})

	if err != nil {
		log.Printf("Error walking content directory: %v", err)
	}

	// 按时间倒序排列
	sort.Slice(Posts, func(i, j int) bool {
		return Posts[i].Date.After(Posts[j].Date)
	})
}

// GetCachedContent 获取渲染后的 HTML，如果不存在则解析并存入缓存
func GetCachedContent(post *Post) string {
	if val, ok := contentCache.Load(post.FilePath); ok {
		return val.(string)
	}

	// 缓存失效或不存在，执行解析
	log.Printf("Cache Miss: Rendering Markdown for %s", post.FilePath)
	freshPost, err := ParseMarkdownFile(post.FilePath)
	if err != nil {
		return "Error rendering content"
	}

	contentCache.Store(post.FilePath, freshPost.Content)
	return freshPost.Content
}

// InvalidateCache 当文件保存时调用
func InvalidateCache(filePath string) {
	contentCache.Delete(filePath)
	log.Printf("Cache Invalidated: %s", filePath)
}

func InitSearchIndex() {
	indexPath := AppConfig.Search.IndexPath
	if Index != nil { Index.Close() }
	os.RemoveAll(indexPath)

	mapping := bleve.NewIndexMapping()
	var err error
	Index, err = bleve.New(indexPath, mapping)
	if err != nil { log.Fatalf("Error creating index: %v", err) }

	for _, post := range Posts {
		Index.Index(post.Slug, post)
	}
}

func SearchPosts(query string) ([]*Post, error) {
	searchQuery := bleve.NewMatchQuery(query)
	searchRequest := bleve.NewSearchRequest(searchQuery)
	results, err := Index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	var foundPosts []*Post
	for _, hit := range results.Hits {
		for _, post := range Posts {
			if post.Slug == hit.ID {
				foundPosts = append(foundPosts, post)
				break
			}
		}
	}
	return foundPosts, nil
}

	


// GetPaginatedPosts 返回分页后的文章列表
func GetPaginatedPosts(page, perPage int) ([]*Post, int) {
	storeLock.RLock()
	defer storeLock.RUnlock()

	total := len(Posts)
	if perPage <= 0 {
		perPage = 10
	}
	totalPages := (total + perPage - 1) / perPage

	start := (page - 1) * perPage
	if start >= total {
		return []*Post{}, totalPages
	}

	end := start + perPage
	if end > total {
		end = total
	}

	return Posts[start:end], totalPages
}


// GetAdjacentPosts 获取上一篇和下一篇文章
func GetAdjacentPosts(currentPost *Post) (prev *Post, next *Post) {
	storeLock.RLock()
	defer storeLock.RUnlock()

	for i, p := range Posts {
		if p.FilePath == currentPost.FilePath {
			// Posts 按时间倒序，所以 i-1 是更新的文章（下一篇），i+1 是更旧的（上一篇）
			if i > 0 {
				next = Posts[i-1]
			}
			if i < len(Posts)-1 {
				prev = Posts[i+1]
			}
			break
		}
	}
	return
}


// TagInfo 标签信息
type TagInfo struct {
	Name      string
	PostCount int
}

// ListTags 获取所有标签及文章数
func ListTags() []TagInfo {
	storeLock.RLock()
	defer storeLock.RUnlock()

	tagMap := make(map[string]int)
	for _, post := range Posts {
		for _, tag := range post.Tags {
			tagMap[tag]++
		}
	}

	tags := make([]TagInfo, 0, len(tagMap))
	for name, count := range tagMap {
		tags = append(tags, TagInfo{Name: name, PostCount: count})
	}
	return tags
}

// GetPostsByTag 获取指定标签的文章
func GetPostsByTag(tag string) []*Post {
	storeLock.RLock()
	defer storeLock.RUnlock()

	var posts []*Post
	for _, post := range Posts {
		for _, t := range post.Tags {
			if t == tag {
				posts = append(posts, post)
				break
			}
		}
	}
	return posts
}


// GetRelatedPosts 获取相关文章（基于标签匹配）
func GetRelatedPosts(currentPost *Post, limit int) []*Post {
	storeLock.RLock()
	defer storeLock.RUnlock()

	if len(currentPost.Tags) == 0 {
		return nil
	}

	// 计算每篇文章与当前文章的标签匹配数
	type scored struct {
		post  *Post
		score int
	}
	var candidates []scored

	currentTags := make(map[string]bool)
	for _, t := range currentPost.Tags {
		currentTags[t] = true
	}

	for _, post := range Posts {
		if post.FilePath == currentPost.FilePath {
			continue
		}
		score := 0
		for _, t := range post.Tags {
			if currentTags[t] {
				score++
			}
		}
		if score > 0 {
			candidates = append(candidates, scored{post, score})
		}
	}

	// 按匹配度降序排序
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			return candidates[i].post.Date.After(candidates[j].post.Date)
		}
		return candidates[i].score > candidates[j].score
	})

	// 取前 limit 篇
	result := make([]*Post, 0, limit)
	for i := 0; i < len(candidates) && i < limit; i++ {
		result = append(result, candidates[i].post)
	}
	return result
}


// GetAllPostsIncludingDrafts 获取所有文章（包括草稿），用于后台管理
func GetAllPostsIncludingDrafts() []*Post {
	storeLock.RLock()
	defer storeLock.RUnlock()

	var allPosts []*Post
	for _, post := range PostsMap {
		allPosts = append(allPosts, post)
	}
	
	// 按时间倒序排列
	sort.Slice(allPosts, func(i, j int) bool {
		return allPosts[i].Date.After(allPosts[j].Date)
	})
	
	return allPosts
}
