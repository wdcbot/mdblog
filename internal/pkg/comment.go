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
	Approved  bool      `json:"approved"`
	ParentID  string    `json:"parent_id,omitempty"` // 父评论ID
	ReplyTo   string    `json:"reply_to,omitempty"`  // 回复的人名
	Replies   []Comment `json:"-"`                   // 子评论，运行时构建
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
func AddComment(postSlug, author, email, content, parentID, replyTo string) (*Comment, error) {
	commentsLock.Lock()
	defer commentsLock.Unlock()

	comment := Comment{
		ID:        time.Now().Format("20060102150405"),
		PostSlug:  postSlug,
		Author:    author,
		Email:     email,
		Content:   content,
		CreatedAt: time.Now(),
		Approved:  true,
		ParentID:  parentID,
		ReplyTo:   replyTo,
	}

	comments = append(comments, comment)
	err := saveComments()
	return &comment, err
}

// GetCommentsByPost 获取文章的评论（树形结构）
func GetCommentsByPost(postSlug string) []Comment {
	commentsLock.RLock()
	defer commentsLock.RUnlock()

	// 获取该文章的所有已审核评论
	var all []*Comment
	for i := range comments {
		if comments[i].PostSlug == postSlug && comments[i].Approved {
			c := comments[i] // 复制一份
			c.Replies = []Comment{}
			all = append(all, &c)
		}
	}

	// 先按时间排序
	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt.Before(all[j].CreatedAt)
	})

	// 建立索引
	commentMap := make(map[string]*Comment)
	for _, c := range all {
		commentMap[c.ID] = c
	}

	// 构建树
	var roots []Comment
	for _, c := range all {
		if c.ParentID == "" {
			roots = append(roots, *c)
		} else if parent, ok := commentMap[c.ParentID]; ok {
			parent.Replies = append(parent.Replies, *c)
		} else {
			// 父评论不存在，作为顶级评论
			roots = append(roots, *c)
		}
	}

	// 重新构建 roots，确保包含最新的 Replies
	var result []Comment
	for _, c := range all {
		if c.ParentID == "" {
			result = append(result, *c)
		}
	}

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
