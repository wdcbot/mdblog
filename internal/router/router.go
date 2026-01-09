package router

import (
	"encoding/xml"
	"fmt"
	"mdblog/internal/pkg"
	"mdblog/internal/theme"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Gzip 压缩
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	// 静态资源缓存中间件
	staticCacheMiddleware := func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=86400") // 缓存 1 天
		c.Next()
	}

	// Static files with cache
	staticGroup := r.Group("/static", staticCacheMiddleware)
	staticGroup.Static("/", filepath.Join("themes", pkg.AppConfig.Theme, "static"))
	r.Static("/admin-static", "admin/static")

	// Frontend routes
	r.GET("/", func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		if page < 1 {
			page = 1
		}
		posts, totalPages := pkg.GetPaginatedPosts(page, pkg.AppConfig.PostsPerPage)
		theme.Render(c, "index.html", gin.H{
			"Posts":       posts,
			"CurrentPage": page,
			"TotalPages":  totalPages,
			"HasPrev":     page > 1,
			"HasNext":     page < totalPages,
			"PrevPage":    page - 1,
			"NextPage":    page + 1,
		})
	})

	r.GET("/categories", func(c *gin.Context) {
		cats, _ := pkg.ListCategories()
		sort.Slice(cats, func(i, j int) bool {
			return cats[i].Name < cats[j].Name
		})
		theme.Render(c, "categories.html", gin.H{
			"Categories": cats,
		})
	})

	r.GET("/category/:name", func(c *gin.Context) {
		name := c.Param("name")
		var posts []*pkg.Post
		for _, p := range pkg.Posts {
			if p.Category == name {
				posts = append(posts, p)
			}
		}
		sort.Slice(posts, func(i, j int) bool {
			if posts[i].Date.Equal(posts[j].Date) {
				return posts[i].Title < posts[j].Title
			}
			return posts[i].Date.After(posts[j].Date)
		})

		theme.Render(c, "category.html", gin.H{
			"Category": name,
			"Posts":    posts,
		})
	})

	// Tags routes
	r.GET("/tags", func(c *gin.Context) {
		tags := pkg.ListTags()
		sort.Slice(tags, func(i, j int) bool {
			return tags[i].PostCount > tags[j].PostCount // 按文章数降序
		})
		theme.Render(c, "tags.html", gin.H{
			"Tags": tags,
		})
	})

	r.GET("/tag/:name", func(c *gin.Context) {
		name := c.Param("name")
		posts := pkg.GetPostsByTag(name)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Date.After(posts[j].Date)
		})
		theme.Render(c, "tag.html", gin.H{
			"Tag":   name,
			"Posts": posts,
		})
	})

	// 1. Pages logic: /page/:slug.html -> /page/*path
	r.GET("/page/*path", func(c *gin.Context) {
		path := c.Param("path") // e.g., "/about.html"

		// Clean the slug: remove leading slash and .html suffix
		slug := strings.TrimPrefix(path, "/")
		slug = strings.TrimSuffix(slug, ".html")

		pages, _ := pkg.ListPages()
		var foundPage *pkg.Page
		for _, p := range pages {
			if p.Slug == slug {
				foundPage = &p
				break
			}
		}

		if foundPage == nil {
			theme.Render(c, "404.html", gin.H{})
			return
		}

		post, err := pkg.ParseMarkdownFile(foundPage.FilePath)
		if err != nil {
			theme.Render(c, "404.html", gin.H{})
			return
		}
		theme.Render(c, "page.html", gin.H{
			"Post":    post,
			"Content": post.Content,
		})
	})

	// 2. Articles logic: /:category/:slug.html -> /:category/*path
	r.GET("/:category/*path", func(c *gin.Context) {
		category := c.Param("category")
		path := c.Param("path") // e.g. "/hello-world.html"

		// Clean the slug
		slug := strings.TrimPrefix(path, "/")
		slug = strings.TrimSuffix(slug, ".html")

		// 构造 key 并查询
		key := strings.ToLower(category + "/" + slug)
		post, ok := pkg.PostsMap[key]

		if !ok {
			// 增加调试信息
			fmt.Printf("[DEBUG] 404 access: RequestPath=%s, ParsedKey=%s\n", c.Request.URL.Path, key)
			// fmt.Println("Available keys in PostsMap:")
			// for k := range pkg.PostsMap {
			// 	fmt.Println(" - ", k)
			// }
			theme.Render(c, "404.html", gin.H{})
			return
		}

		content := pkg.GetCachedContent(post)
		pkg.RecordView(post.Slug) // 记录访问
		prev, next := pkg.GetAdjacentPosts(post)
		related := pkg.GetRelatedPosts(post, 3)
		comments := pkg.GetCommentsByPost(post.Slug)
		theme.Render(c, "post.html", gin.H{
			"Post":         post,
			"Content":      content,
			"PrevPost":     prev,
			"NextPost":     next,
			"RelatedPosts": related,
			"Comments":     comments,
		})
	})

	r.GET("/search", func(c *gin.Context) {
		query := c.Query("q")
		foundPosts, _ := pkg.SearchPosts(query)
		theme.Render(c, "search.html", gin.H{
			"Posts": foundPosts,
			"Query": query,
		})
	})

	// RSS Feed
	r.GET("/feed.xml", func(c *gin.Context) {
		c.Header("Content-Type", "application/rss+xml; charset=utf-8")
		feed := generateRSSFeed()
		c.String(http.StatusOK, feed)
	})

	// Sitemap
	r.GET("/sitemap.xml", func(c *gin.Context) {
		c.Header("Content-Type", "application/xml; charset=utf-8")
		sitemap := generateSitemap()
		c.String(http.StatusOK, sitemap)
	})

	// Robots.txt
	r.GET("/robots.txt", func(c *gin.Context) {
		robots := fmt.Sprintf("User-agent: *\nAllow: /\nSitemap: %s/sitemap.xml\n", pkg.AppConfig.Site.BaseURL)
		c.String(http.StatusOK, robots)
	})

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":     "ok",
			"posts":      len(pkg.Posts),
			"categories": len(pkg.PostsMap),
		})
	})

	// Comment submission
	r.POST("/comment", func(c *gin.Context) {
		postSlug := c.PostForm("post_slug")
		author := c.PostForm("author")
		email := c.PostForm("email")
		content := c.PostForm("content")

		// 简单验证
		if postSlug == "" || author == "" || content == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必填字段"})
			return
		}
		if len(content) > 2000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "评论内容过长"})
			return
		}

		_, err := pkg.AddComment(postSlug, author, email, content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "评论已提交，等待审核"})
	})

	// Admin routes
	admin := r.Group("/admin", gin.BasicAuth(gin.Accounts{
		pkg.AppConfig.AdminUsername: pkg.AppConfig.AdminPassword,
	}))

	admin.GET("/", func(c *gin.Context) {
		cats, _ := pkg.ListCategories()
		pages, _ := pkg.ListPages()
		stats := pkg.GetStats()
		chartLabels, chartValues := pkg.GetDailyViewsChart()
		pendingComments := pkg.GetPendingComments()

		recentPosts := pkg.Posts
		if len(recentPosts) > 5 {
			recentPosts = recentPosts[:5]
		}

		err := theme.AdminTemplates.ExecuteTemplate(c.Writer, "admin-index.html", gin.H{
			"RecentPosts":     recentPosts,
			"Categories":      cats,
			"Pages":           pages,
			"PostCount":       len(pkg.Posts),
			"Stats":           stats,
			"ChartLabels":     chartLabels,
			"ChartValues":     chartValues,
			"PendingComments": len(pendingComments),
			"Tab":             "overview",
		})
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
		}
	})

	// New Route: Manage Posts
	admin.GET("/posts", func(c *gin.Context) {
		cats, _ := pkg.ListCategories()
		allPosts := pkg.GetAllPostsIncludingDrafts()
		err := theme.AdminTemplates.ExecuteTemplate(c.Writer, "admin-posts.html", gin.H{
			"Posts":      allPosts,
			"Categories": cats,
			"Tab":        "posts",
		})
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
		}
	})

	admin.GET("/categories", func(c *gin.Context) {
		cats, err := pkg.ListCategories()
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		err = theme.AdminTemplates.ExecuteTemplate(c.Writer, "admin-categories.html", gin.H{
			"Categories": cats,
			"Tab":        "categories",
		})
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
		}
	})

	admin.POST("/categories/create", func(c *gin.Context) {
		name := c.PostForm("name")
		if err := pkg.CreateCategory(name); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	admin.POST("/categories/rename", func(c *gin.Context) {
		oldName := c.PostForm("old_name")
		newName := c.PostForm("new_name")
		if err := pkg.RenameCategory(oldName, newName); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		pkg.LoadAllPosts()
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	admin.POST("/categories/delete", func(c *gin.Context) {
		name := c.PostForm("name")
		if err := pkg.DeleteCategory(name); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Page Management
	admin.GET("/pages", func(c *gin.Context) {
		pages, _ := pkg.ListPages()
		err := theme.AdminTemplates.ExecuteTemplate(c.Writer, "admin-pages.html", gin.H{
			"Pages": pages,
			"Tab":   "pages",
		})
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
		}
	})

	admin.POST("/pages/create", func(c *gin.Context) {
		title := c.PostForm("title")
		path, err := pkg.CreatePage(title)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "path": path})
	})

	admin.POST("/pages/delete", func(c *gin.Context) {
		slug := c.PostForm("slug")
		if err := pkg.DeletePage(slug); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	admin.POST("/create", func(c *gin.Context) {
		title := c.PostForm("title")
		category := c.PostForm("category")
		slug := c.PostForm("slug")
		draft := c.PostForm("draft") == "true"
		path, err := pkg.CreatePostFile(category, title, slug, draft)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		pkg.LoadAllPosts()
		pkg.InitSearchIndex()
		c.JSON(http.StatusOK, gin.H{"status": "ok", "path": path})
	})

	admin.POST("/delete", func(c *gin.Context) {
		path := c.PostForm("path")
		// 安全校验：防止目录遍历
		if !pkg.IsPathSafe(path) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if err := pkg.DeletePostFile(path); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		pkg.LoadAllPosts()
		pkg.InitSearchIndex()

		// Support HTMX: Return empty content to remove the element from DOM
		if c.GetHeader("HX-Request") == "true" {
			c.Status(http.StatusOK)
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 批量删除
	admin.POST("/batch-delete", func(c *gin.Context) {
		var req struct {
			Paths []string `json:"paths"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		for _, path := range req.Paths {
			if !pkg.IsPathSafe(path) {
				continue
			}
			pkg.DeletePostFile(path)
		}

		pkg.LoadAllPosts()
		pkg.InitSearchIndex()
		c.JSON(http.StatusOK, gin.H{"status": "ok", "deleted": len(req.Paths)})
	})

	admin.GET("/edit", func(c *gin.Context) {
		path := c.Query("path")
		// 安全校验：防止目录遍历
		if !pkg.IsPathSafe(path) {
			c.String(http.StatusForbidden, "Access denied")
			return
		}
		content, err := pkg.ReadPostFile(path)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		err = theme.AdminTemplates.ExecuteTemplate(c.Writer, "admin-edit.html", gin.H{
			"Path":    path,
			"Content": content,
		})
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
		}
	})

	admin.POST("/save", func(c *gin.Context) {
		path := c.PostForm("path")
		content := c.PostForm("content")
		// 安全校验：防止目录遍历
		if !pkg.IsPathSafe(path) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		err := pkg.SavePostFile(path, content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		pkg.InvalidateCache(path)
		pkg.LoadAllPosts()
		pkg.InitSearchIndex()
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	admin.GET("/settings", func(c *gin.Context) {
		themes, _ := pkg.ListThemes()
		err := theme.AdminTemplates.ExecuteTemplate(c.Writer, "admin-settings.html", gin.H{
			"Config": pkg.AppConfig,
			"Themes": themes,
			"Tab":    "settings",
		})
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
		}
	})

	// Comment Management
	admin.GET("/comments", func(c *gin.Context) {
		comments := pkg.GetAllComments()
		pending := pkg.GetPendingComments()
		err := theme.AdminTemplates.ExecuteTemplate(c.Writer, "admin-comments.html", gin.H{
			"Comments":     comments,
			"PendingCount": len(pending),
			"Tab":          "comments",
		})
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
		}
	})

	admin.POST("/comments/approve", func(c *gin.Context) {
		id := c.PostForm("id")
		pkg.ApproveComment(id)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	admin.POST("/comments/delete", func(c *gin.Context) {
		id := c.PostForm("id")
		pkg.DeleteComment(id)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	admin.POST("/settings/update", func(c *gin.Context) {
		siteTitle := c.PostForm("site_title")
		siteDesc := c.PostForm("site_desc")
		baseURL := c.PostForm("base_url")
		themeName := c.PostForm("theme")
		postsPerPage, _ := strconv.Atoi(c.PostForm("posts_per_page"))

		if err := pkg.UpdateConfig(siteTitle, siteDesc, baseURL, themeName, postsPerPage); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		// Reload theme templates in case theme changed
		theme.InitPongo2()

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}

// RSS Feed 结构
type RSS struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel RSSChannel `xml:"channel"`
}

type RSSChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []RSSItem `xml:"item"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

func generateRSSFeed() string {
	items := make([]RSSItem, 0, 20)
	posts := pkg.Posts
	if len(posts) > 20 {
		posts = posts[:20]
	}

	for _, post := range posts {
		link := fmt.Sprintf("%s/%s/%s.html", pkg.AppConfig.Site.BaseURL, post.Category, post.Slug)
		items = append(items, RSSItem{
			Title:       post.Title,
			Link:        link,
			Description: post.Summary,
			PubDate:     post.Date.Format(time.RFC1123Z),
			GUID:        link,
		})
	}

	rss := RSS{
		Version: "2.0",
		Channel: RSSChannel{
			Title:       pkg.AppConfig.Site.Title,
			Link:        pkg.AppConfig.Site.BaseURL,
			Description: pkg.AppConfig.Site.Description,
			Items:       items,
		},
	}

	output, _ := xml.MarshalIndent(rss, "", "  ")
	return xml.Header + string(output)
}


// Sitemap 结构
type URLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []SitemapURL `xml:"url"`
}

type SitemapURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

func generateSitemap() string {
	urls := []SitemapURL{
		{Loc: pkg.AppConfig.Site.BaseURL, ChangeFreq: "daily", Priority: "1.0"},
	}

	// 添加所有文章
	for _, post := range pkg.Posts {
		urls = append(urls, SitemapURL{
			Loc:        fmt.Sprintf("%s/%s/%s.html", pkg.AppConfig.Site.BaseURL, post.Category, post.Slug),
			LastMod:    post.Date.Format("2006-01-02"),
			ChangeFreq: "weekly",
			Priority:   "0.8",
		})
	}

	// 添加分类页
	cats, _ := pkg.ListCategories()
	for _, cat := range cats {
		urls = append(urls, SitemapURL{
			Loc:        fmt.Sprintf("%s/category/%s", pkg.AppConfig.Site.BaseURL, cat.Name),
			ChangeFreq: "weekly",
			Priority:   "0.6",
		})
	}

	// 添加标签页
	tags := pkg.ListTags()
	for _, tag := range tags {
		urls = append(urls, SitemapURL{
			Loc:        fmt.Sprintf("%s/tag/%s", pkg.AppConfig.Site.BaseURL, tag.Name),
			ChangeFreq: "weekly",
			Priority:   "0.5",
		})
	}

	sitemap := URLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	output, _ := xml.MarshalIndent(sitemap, "", "  ")
	return xml.Header + string(output)
}
