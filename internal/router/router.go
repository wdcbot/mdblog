package router

import (
	"archive/zip"
	"crypto/rand"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"mdblog/internal/pkg"
	"mdblog/internal/theme"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// Session 管理
var (
	sessions     = make(map[string]time.Time)
	sessionsLock sync.RWMutex
)

func generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func isValidSession(sessionID string) bool {
	sessionsLock.RLock()
	defer sessionsLock.RUnlock()
	if exp, ok := sessions[sessionID]; ok {
		return time.Now().Before(exp)
	}
	return false
}

func createSession() string {
	return createSessionWithExpiry(24 * time.Hour)
}

func createSessionWithExpiry(expiry time.Duration) string {
	sessionsLock.Lock()
	defer sessionsLock.Unlock()
	sessionID := generateSessionID()
	sessions[sessionID] = time.Now().Add(expiry)
	return sessionID
}

func deleteSession(sessionID string) {
	sessionsLock.Lock()
	defer sessionsLock.Unlock()
	delete(sessions, sessionID)
}

// Admin 认证中间件
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("admin_session")
		if err != nil || !isValidSession(sessionID) {
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

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
			if posts[i].Pinned != posts[j].Pinned {
				return posts[i].Pinned
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
		parentID := c.PostForm("parent_id")
		replyTo := c.PostForm("reply_to")

		// 简单验证
		if postSlug == "" || author == "" || content == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必填字段"})
			return
		}
		if len(content) > 2000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "评论内容过长"})
			return
		}

		_, err := pkg.AddComment(postSlug, author, email, content, parentID, replyTo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "评论成功"})
	})

	// Admin login page
	r.GET("/admin/login", func(c *gin.Context) {
		// 已登录则跳转
		if sessionID, err := c.Cookie("admin_session"); err == nil && isValidSession(sessionID) {
			c.Redirect(http.StatusFound, "/admin/")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>登录 - mdblog Admin</title>
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
	<style>
		* { box-sizing: border-box; }
		body { font-family: system-ui, -apple-system, sans-serif; display: flex; align-items: center; justify-content: center; min-height: 100vh; margin: 0; background: linear-gradient(135deg, #1e3a5f 0%, #0f172a 100%); padding: 1rem; }
		.login-box { width: 100%; max-width: 400px; padding: 2.5rem; background: #fff; border-radius: 16px; box-shadow: 0 20px 60px rgba(0,0,0,0.3); }
		.login-header { text-align: center; margin-bottom: 2rem; }
		.login-header h1 { margin: 0 0 0.5rem; font-size: 1.75rem; color: #111827; font-weight: 700; }
		.login-header p { margin: 0; color: #6b7280; font-size: 0.95rem; }
		.form-group { margin-bottom: 1.25rem; }
		.form-group label { display: block; margin-bottom: 0.5rem; font-weight: 500; color: #374151; font-size: 0.9rem; }
		.form-group input[type="text"], .form-group input[type="password"] { width: 100%; padding: 0.875rem 1rem; border: 1px solid #d1d5db; border-radius: 10px; font-size: 1rem; transition: all 0.2s; }
		.form-group input:focus { outline: none; border-color: #2563eb; box-shadow: 0 0 0 3px rgba(37,99,235,0.1); }
		.form-options { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; flex-wrap: wrap; gap: 0.5rem; }
		.form-options label { display: flex; align-items: center; gap: 0.5rem; cursor: pointer; font-weight: 400; color: #6b7280; font-size: 0.9rem; }
		.form-options input[type="checkbox"] { width: auto; }
		.btn-login { width: 100%; padding: 0.875rem; background: #2563eb; color: #fff; border: none; border-radius: 10px; font-size: 1rem; font-weight: 600; cursor: pointer; transition: background 0.2s; }
		.btn-login:hover { background: #1d4ed8; }
		.btn-login:disabled { background: #9ca3af; cursor: not-allowed; }
		.error-msg { background: #fef2f2; color: #dc2626; padding: 0.75rem 1rem; border-radius: 8px; margin-bottom: 1rem; font-size: 0.9rem; display: none; }
		.back-link { display: block; text-align: center; margin-top: 1.5rem; color: #6b7280; text-decoration: none; font-size: 0.9rem; }
		.back-link:hover { color: #2563eb; }
		@media (max-width: 480px) {
			.login-box { padding: 1.5rem; }
			.login-header h1 { font-size: 1.5rem; }
			.form-options { flex-direction: column; align-items: flex-start; }
		}
	</style>
</head>
<body>
	<div class="login-box">
		<div class="login-header">
			<h1><i class="fa-solid fa-feather"></i> mdblog</h1>
			<p>后台管理系统</p>
		</div>
		<div class="error-msg" id="error-msg"></div>
		<form id="login-form">
			<div class="form-group">
				<label>用户名</label>
				<input type="text" name="username" id="username" required autocomplete="username">
			</div>
			<div class="form-group">
				<label>密码</label>
				<input type="password" name="password" id="password" required autocomplete="current-password">
			</div>
			<div class="form-options">
				<label><input type="checkbox" name="remember_account" id="remember_account"> 记住账号</label>
				<label><input type="checkbox" name="remember" id="remember_session"> 7天免登录</label>
			</div>
			<button type="submit" class="btn-login">登录</button>
		</form>
		<a href="/" class="back-link">← 返回网站首页</a>
	</div>
	<script>
		const usernameInput = document.getElementById('username');
		const passwordInput = document.getElementById('password');
		const rememberAccount = document.getElementById('remember_account');
		
		// 从 localStorage 恢复账号信息
		const savedUsername = localStorage.getItem('admin_username');
		const savedPassword = localStorage.getItem('admin_password');
		if (savedUsername) {
			usernameInput.value = savedUsername;
			rememberAccount.checked = true;
		}
		if (savedPassword) {
			passwordInput.value = atob(savedPassword); // base64 解码
		}
		
		document.getElementById('login-form').addEventListener('submit', function(e) {
			e.preventDefault();
			const btn = this.querySelector('button');
			const errMsg = document.getElementById('error-msg');
			btn.disabled = true;
			btn.textContent = '登录中...';
			errMsg.style.display = 'none';
			
			// 保存或清除账号信息
			if (rememberAccount.checked) {
				localStorage.setItem('admin_username', usernameInput.value);
				localStorage.setItem('admin_password', btoa(passwordInput.value)); // base64 编码
			} else {
				localStorage.removeItem('admin_username');
				localStorage.removeItem('admin_password');
			}
			
			fetch('/admin/login', {
				method: 'POST',
				body: new FormData(this)
			})
			.then(res => res.json())
			.then(data => {
				if (data.status === 'ok') {
					window.location.href = '/admin/';
				} else {
					errMsg.textContent = data.error || '登录失败';
					errMsg.style.display = 'block';
					btn.disabled = false;
					btn.textContent = '登录';
				}
			})
			.catch(() => {
				errMsg.textContent = '网络错误，请重试';
				errMsg.style.display = 'block';
				btn.disabled = false;
				btn.textContent = '登录';
			});
		});
	</script>
</body>
</html>`))
	})

	// Admin login POST
	r.POST("/admin/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")
		remember := c.PostForm("remember") == "on"

		if username == pkg.AppConfig.AdminUsername && password == pkg.AppConfig.AdminPassword {
			var expiry time.Duration
			var maxAge int
			if remember {
				expiry = 7 * 24 * time.Hour // 7天
				maxAge = 7 * 86400
			} else {
				expiry = 24 * time.Hour // 1天
				maxAge = 86400
			}
			sessionID := createSessionWithExpiry(expiry)
			c.SetCookie("admin_session", sessionID, maxAge, "/", "", false, true)
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		}
	})

	// Admin logout
	r.GET("/admin/logout", func(c *gin.Context) {
		if sessionID, err := c.Cookie("admin_session"); err == nil {
			deleteSession(sessionID)
		}
		c.SetCookie("admin_session", "", -1, "/", "", false, true)
		c.Redirect(http.StatusFound, "/admin/login")
	})

	// Admin routes (protected)
	admin := r.Group("/admin", AdminAuthMiddleware())

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

	// 基础设置
	admin.POST("/settings/update", func(c *gin.Context) {
		siteTitle := c.PostForm("site_title")
		siteDesc := c.PostForm("site_desc")
		baseURL := c.PostForm("base_url")
		themeName := c.PostForm("theme")
		postsPerPage, _ := strconv.Atoi(c.PostForm("posts_per_page"))
		keywords := c.PostForm("keywords")
		author := c.PostForm("author")

		if err := pkg.UpdateBasicConfig(siteTitle, siteDesc, baseURL, themeName, postsPerPage, keywords, author); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		// Reload theme templates in case theme changed
		theme.InitPongo2()

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 外观设置
	admin.POST("/settings/appearance", func(c *gin.Context) {
		heroTitle := c.PostForm("hero_title")
		defaultTheme := c.PostForm("default_theme")
		accentColor := c.PostForm("accent_color")
		favicon := c.PostForm("favicon")
		logo := c.PostForm("logo")
		logoHeight, _ := strconv.Atoi(c.PostForm("logo_height"))
		if logoHeight <= 0 {
			logoHeight = 36 // 默认高度
		}

		if err := pkg.UpdateAppearanceConfig(heroTitle, defaultTheme, accentColor, favicon, logo, logoHeight); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 功能开关
	admin.POST("/settings/features", func(c *gin.Context) {
		commentsEnabled := c.PostForm("comments_enabled") == "true"
		tocEnabled := c.PostForm("toc_enabled") == "true"
		readingTimeEnabled := c.PostForm("reading_time_enabled") == "true"

		if err := pkg.UpdateFeaturesConfig(commentsEnabled, tocEnabled, readingTimeEnabled); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 页脚设置
	admin.POST("/settings/footer", func(c *gin.Context) {
		copyright := c.PostForm("copyright")
		icp := c.PostForm("icp")
		linkNames := c.PostFormArray("link_name[]")
		linkURLs := c.PostFormArray("link_url[]")
		linkIcons := c.PostFormArray("link_icon[]")

		var links []pkg.FooterLink
		for i := 0; i < len(linkNames); i++ {
			if linkNames[i] != "" && linkURLs[i] != "" {
				links = append(links, pkg.FooterLink{
					Name: linkNames[i],
					URL:  linkURLs[i],
					Icon: linkIcons[i],
				})
			}
		}

		if err := pkg.UpdateFooterConfig(copyright, icp, links); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 统计代码
	admin.POST("/settings/analytics", func(c *gin.Context) {
		analytics := c.PostForm("analytics")

		if err := pkg.UpdateAnalyticsConfig(analytics); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 图片上传
	admin.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请选择文件"})
			return
		}

		// 验证文件类型
		ext := strings.ToLower(filepath.Ext(file.Filename))
		allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".svg": true}
		if !allowedExts[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件类型"})
			return
		}

		// 限制文件大小 (10MB)
		if file.Size > 10*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "文件大小不能超过10MB"})
			return
		}

		// 创建上传目录
		uploadDir := filepath.Join("themes", pkg.AppConfig.Theme, "static", "uploads")
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建目录失败"})
			return
		}

		// 生成唯一文件名
		filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		savePath := filepath.Join(uploadDir, filename)

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
			return
		}

		// 返回访问URL
		url := fmt.Sprintf("/static/uploads/%s", filename)
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"url":    url,
			"name":   file.Filename,
		})
	})

	// Markdown 预览
	admin.POST("/preview", func(c *gin.Context) {
		content := c.PostForm("content")
		html := pkg.RenderMarkdownPreview(content)
		c.JSON(http.StatusOK, gin.H{"html": html})
	})

	// 数据备份 - 导出
	admin.GET("/backup", func(c *gin.Context) {
		// 创建临时 zip 文件
		tmpFile, err := os.CreateTemp("", "mdblog-backup-*.zip")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建备份失败"})
			return
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		zipWriter := zip.NewWriter(tmpFile)

		// 备份 content 目录
		err = addDirToZip(zipWriter, "content", "content")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "备份内容失败"})
			return
		}

		// 备份 data 目录
		err = addDirToZip(zipWriter, "data", "data")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "备份数据失败"})
			return
		}

		// 备份 config.yaml
		err = addFileToZip(zipWriter, "config.yaml", "config.yaml")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "备份配置失败"})
			return
		}

		zipWriter.Close()

		// 发送文件
		filename := fmt.Sprintf("mdblog-backup-%s.zip", time.Now().Format("20060102-150405"))
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Header("Content-Type", "application/zip")
		c.File(tmpFile.Name())
	})

	// 数据恢复 - 导入
	admin.POST("/restore", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请选择备份文件"})
			return
		}

		// 验证文件类型
		if !strings.HasSuffix(file.Filename, ".zip") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请上传 .zip 格式的备份文件"})
			return
		}

		// 保存上传的文件
		tmpFile, err := os.CreateTemp("", "mdblog-restore-*.zip")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "处理文件失败"})
			return
		}
		defer os.Remove(tmpFile.Name())

		if err := c.SaveUploadedFile(file, tmpFile.Name()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
			return
		}

		// 解压文件
		if err := extractZip(tmpFile.Name(), "."); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "解压失败: " + err.Error()})
			return
		}

		// 重新加载数据
		pkg.InitConfig()
		pkg.LoadAllPosts()
		pkg.InitSearchIndex()
		pkg.LoadComments()
		pkg.LoadStats()

		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "数据恢复成功"})
	})

	return r
}

// 添加目录到 zip
func addDirToZip(zipWriter *zip.Writer, srcDir, baseInZip string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算在 zip 中的路径
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		zipPath := filepath.Join(baseInZip, relPath)

		if info.IsDir() {
			_, err := zipWriter.Create(zipPath + "/")
			return err
		}

		return addFileToZip(zipWriter, path, zipPath)
	})
}

// 添加文件到 zip
func addFileToZip(zipWriter *zip.Writer, srcFile, zipPath string) error {
	file, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer file.Close()

	writer, err := zipWriter.Create(zipPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

// 解压 zip 文件
func extractZip(zipFile, destDir string) error {
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		// 安全检查：防止 zip slip 攻击
		destPath := filepath.Join(destDir, file.Name)
		if !strings.HasPrefix(destPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("非法文件路径: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(destPath, 0755)
			continue
		}

		// 确保父目录存在
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		// 解压文件
		srcFile, err := file.Open()
		if err != nil {
			return err
		}

		destFile, err := os.Create(destPath)
		if err != nil {
			srcFile.Close()
			return err
		}

		_, err = io.Copy(destFile, srcFile)
		srcFile.Close()
		destFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
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
