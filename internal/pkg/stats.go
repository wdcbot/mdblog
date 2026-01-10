package pkg

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Stats 访问统计
type Stats struct {
	TotalViews   int            `json:"total_views"`
	TodayViews   int            `json:"today_views"`
	DailyViews   map[string]int `json:"daily_views"` // 最近 7 天
	TopPosts     map[string]int `json:"top_posts"`   // 热门文章
	LastUpdated  string         `json:"last_updated"`
}

var (
	stats     Stats
	statsLock sync.RWMutex
	statsFile = "data/stats.json"
)

// InitStats 初始化统计
func InitStats() {
	os.MkdirAll("data", 0755)
	loadStats()
}

func loadStats() {
	statsLock.Lock()
	defer statsLock.Unlock()

	stats = Stats{
		DailyViews: make(map[string]int),
		TopPosts:   make(map[string]int),
	}

	data, err := os.ReadFile(statsFile)
	if err != nil {
		return
	}
	json.Unmarshal(data, &stats)

	// 检查是否需要重置今日统计
	today := time.Now().Format("2006-01-02")
	if stats.LastUpdated != today {
		stats.TodayViews = 0
		stats.LastUpdated = today
	}
}

func saveStats() error {
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(statsFile, data, 0644)
}

// RecordView 记录访问
func RecordView(postSlug string) {
	statsLock.Lock()
	defer statsLock.Unlock()

	today := time.Now().Format("2006-01-02")

	// 更新日期
	if stats.LastUpdated != today {
		stats.TodayViews = 0
		stats.LastUpdated = today
	}

	stats.TotalViews++
	stats.TodayViews++

	// 记录每日访问
	if stats.DailyViews == nil {
		stats.DailyViews = make(map[string]int)
	}
	stats.DailyViews[today]++

	// 清理超过 7 天的数据
	cutoff := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	for date := range stats.DailyViews {
		if date < cutoff {
			delete(stats.DailyViews, date)
		}
	}

	// 记录文章访问
	if postSlug != "" {
		if stats.TopPosts == nil {
			stats.TopPosts = make(map[string]int)
		}
		stats.TopPosts[postSlug]++
	}

	saveStats()
}

// GetStats 获取统计数据
func GetStats() Stats {
	statsLock.RLock()
	defer statsLock.RUnlock()
	return stats
}

// GetDailyViewsChart 获取最近 7 天的访问数据（用于图表）
func GetDailyViewsChart() ([]string, []int) {
	statsLock.RLock()
	defer statsLock.RUnlock()

	var labels []string
	var values []int

	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		labels = append(labels, time.Now().AddDate(0, 0, -i).Format("01-02"))
		values = append(values, stats.DailyViews[date])
	}

	return labels, values
}

// LoadStats 重新加载统计数据（用于数据恢复后）
func LoadStats() {
	loadStats()
}
