package pkg

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
	"time"
)

// Comment 评论结构
type Comment struct {
	ID        string    `json:"id"`
	PostSlug  string    `json:"post_slug"`
	Author    string    `json:"author"`
	Email     string    `json:"email"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Approved  bool      `json:"approved"` // 是否审核通过
}

var (
	comments     []Comment
	commentsLock sync.RWMutex
	commentsFile = "data/comments.json"
)

// InitComments 初始化评论系统
func InitComments() {
	os.MkdirAll("data", 0755)
	LoadComments()
}

// LoadComments 从文件加载评论
func LoadComments() {
	commentsLock.Lock()
	defer commentsLock.Unlock()

	comments = []Comment{}
	data, err := os.ReadFile(commentsFile)
	if err != nil {
		return
	}
	json.Unmarshal(data, &comments)
}

// SaveComments 保存评论到文件
func saveComments() error {
	data, err := json.MarshalIndent(comments, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(commentsFile, data, 0644)
}

// AddComment 添加评论
func AddComment(postSlug, author, email, content string) (*Comment, error) {
	commentsLock.Lock()
	defer commentsLock.Unlock()

	comment := Comment{
		ID:        time.Now().Format("20060102150405"),
		PostSlug:  postSlug,
		Author:    author,
		Email:     email,
		Content:   content,
		CreatedAt: time.Now(),
		Approved:  false, // 默认需要审核
	}

	comments = append(comments, comment)
	err := saveComments()
	return &comment, err
}

// GetCommentsByPost 获取文章的已审核评论
func GetCommentsByPost(postSlug string) []Comment {
	commentsLock.RLock()
	defer commentsLock.RUnlock()

	var result []Comment
	for _, c := range comments {
		if c.PostSlug == postSlug && c.Approved {
			result = append(result, c)
		}
	}
	
	// 按时间正序
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result
}

// GetAllComments 获取所有评论（后台用）
func GetAllComments() []Comment {
	commentsLock.RLock()
	defer commentsLock.RUnlock()

	result := make([]Comment, len(comments))
	copy(result, comments)
	
	// 按时间倒序
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	return result
}

// GetPendingComments 获取待审核评论
func GetPendingComments() []Comment {
	commentsLock.RLock()
	defer commentsLock.RUnlock()

	var result []Comment
	for _, c := range comments {
		if !c.Approved {
			result = append(result, c)
		}
	}
	
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	return result
}

// ApproveComment 审核通过评论
func ApproveComment(id string) error {
	commentsLock.Lock()
	defer commentsLock.Unlock()

	for i := range comments {
		if comments[i].ID == id {
			comments[i].Approved = true
			return saveComments()
		}
	}
	return nil
}

// DeleteComment 删除评论
func DeleteComment(id string) error {
	commentsLock.Lock()
	defer commentsLock.Unlock()

	for i := range comments {
		if comments[i].ID == id {
			comments = append(comments[:i], comments[i+1:]...)
			return saveComments()
		}
	}
	return nil
}
