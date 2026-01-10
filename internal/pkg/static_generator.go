package pkg

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flosch/pongo2/v6"
)

// StaticGenerator 静态站点生成器
type StaticGenerator struct {
	OutputDir string
	BaseURL   string
	templates *pongo2.TemplateSet
}

// NewStaticGenerator 创建生成器
func NewStaticGenerator(outputDir string) *StaticGenerator {
	loader := pongo2.MustNewLocalFileSystemLoader(filepath.Join("themes", AppConfig.Theme, "layouts"))
	tplSet := pongo2.NewSet("static", loader)

	return &StaticGenerator{
		OutputDir: outputDir,
		BaseURL:   AppConfig.Site.BaseURL,
		templates: tplSet,
	}
}

// Generate 生成静态站点
func (g *StaticGenerator) Generate() error {
	log.Println("🚀 开始生成静态站点...")

	// 清理输出目录
	os.RemoveAll(g.OutputDir)
	os.MkdirAll(g.OutputDir, 0755)

	// 复制静态资源
	if err := g.copyStatic(); err != nil {
		return fmt.Errorf("复制静态资源失败: %v", err)
	}

	// 生成首页
	if err := g.generateIndex(); err != nil {
		return fmt.Errorf("生成首页失败: %v", err)
	}

	// 生成文章页
	if err := g.generatePosts(); err != nil {
		return fmt.Errorf("生成文章页失败: %v", err)
	}

	// 生成分类页
	if err := g.generateCategories(); err != nil {
		return fmt.Errorf("生成分类页失败: %v", err)
	}

	// 生成标签页
	if err := g.generateTags(); err != nil {
		return fmt.Errorf("生成标签页失败: %v", err)
	}

	// 生成独立页面
	if err := g.generatePages(); err != nil {
		return fmt.Errorf("生成独立页面失败: %v", err)
	}

	// 生成 RSS
	if err := g.generateRSS(); err != nil {
		return fmt.Errorf("生成 RSS 失败: %v", err)
	}

	// 生成 sitemap
	if err := g.generateSitemap(); err != nil {
		return fmt.Errorf("生成 sitemap 失败: %v", err)
	}

	log.Println("✅ 静态站点生成完成！输出目录:", g.OutputDir)
	return nil
}

func (g *StaticGenerator) copyStatic() error {
	log.Println("📁 复制静态资源...")

	// 复制主题静态文件
	srcDir := filepath.Join("themes", AppConfig.Theme, "static")
	dstDir := filepath.Join(g.OutputDir, "static")
	if err := copyDir(srcDir, dstDir); err != nil {
		return err
	}

	// 复制上传的图片
	if _, err := os.Stat("uploads"); err == nil {
		if err := copyDir("uploads", filepath.Join(g.OutputDir, "uploads")); err != nil {
			return err
		}
	}

	return nil
}

func (g *StaticGenerator) generateIndex() error {
	log.Println("📄 生成首页...")

	posts, totalPages := GetPaginatedPosts(1, AppConfig.PostsPerPage)
	if totalPages == 0 {
		totalPages = 1
	}

	// 生成所有分页
	for page := 1; page <= totalPages; page++ {
		pagePosts, _ := GetPaginatedPosts(page, AppConfig.PostsPerPage)

		ctx := g.baseContext()
		ctx["Posts"] = pagePosts
		ctx["CurrentPage"] = page
		ctx["TotalPages"] = totalPages
		ctx["HasPrev"] = page > 1
		ctx["HasNext"] = page < totalPages
		ctx["PrevPage"] = page - 1
		ctx["NextPage"] = page + 1

		var outPath string
		if page == 1 {
			outPath = filepath.Join(g.OutputDir, "index.html")
		} else {
			os.MkdirAll(filepath.Join(g.OutputDir, "page", fmt.Sprintf("%d", page)), 0755)
			outPath = filepath.Join(g.OutputDir, "page", fmt.Sprintf("%d", page), "index.html")
		}

		if err := g.renderTemplate("index.html", ctx, outPath); err != nil {
			return err
		}
	}

	_ = posts // 避免未使用警告
	return nil
}

func (g *StaticGenerator) generatePosts() error {
	log.Println("📝 生成文章页...")

	// 获取所有文章
	allPosts, _ := GetPaginatedPosts(1, 10000)
	
	for i, post := range allPosts {
		content := RenderMarkdownPreview(post.Content)

		ctx := g.baseContext()
		ctx["Post"] = post
		ctx["Content"] = content

		// 上一篇/下一篇
		if i > 0 {
			ctx["NextPost"] = allPosts[i-1]
		}
		if i < len(allPosts)-1 {
			ctx["PrevPost"] = allPosts[i+1]
		}

		// 评论（静态版本为空）
		ctx["Comments"] = []Comment{}

		// 输出路径
		outDir := filepath.Join(g.OutputDir, post.Category)
		os.MkdirAll(outDir, 0755)
		outPath := filepath.Join(outDir, post.Slug+".html")

		if err := g.renderTemplate("post.html", ctx, outPath); err != nil {
			return err
		}
	}

	return nil
}

func (g *StaticGenerator) generateCategories() error {
	log.Println("📂 生成分类页...")

	categories, _ := ListCategories()

	// 分类列表页
	ctx := g.baseContext()
	ctx["Categories"] = categories
	os.MkdirAll(filepath.Join(g.OutputDir, "categories"), 0755)
	if err := g.renderTemplate("categories.html", ctx, filepath.Join(g.OutputDir, "categories", "index.html")); err != nil {
		return err
	}

	// 每个分类的文章列表
	for _, cat := range categories {
		posts := getPostsByCategory(cat.Name)

		ctx := g.baseContext()
		ctx["Category"] = cat
		ctx["Posts"] = posts

		outDir := filepath.Join(g.OutputDir, "category", cat.Name)
		os.MkdirAll(outDir, 0755)
		if err := g.renderTemplate("category.html", ctx, filepath.Join(outDir, "index.html")); err != nil {
			return err
		}
	}

	return nil
}

// getPostsByCategory 获取分类下的文章
func getPostsByCategory(category string) []*Post {
	allPosts, _ := GetPaginatedPosts(1, 10000)
	var result []*Post
	for _, p := range allPosts {
		if p.Category == category {
			result = append(result, p)
		}
	}
	return result
}

func (g *StaticGenerator) generateTags() error {
	log.Println("🏷️ 生成标签页...")

	tags := ListTags()

	// 标签列表页
	ctx := g.baseContext()
	ctx["Tags"] = tags
	os.MkdirAll(filepath.Join(g.OutputDir, "tags"), 0755)
	if err := g.renderTemplate("tags.html", ctx, filepath.Join(g.OutputDir, "tags", "index.html")); err != nil {
		return err
	}

	// 每个标签的文章列表
	for _, tag := range tags {
		posts := GetPostsByTag(tag.Name)

		ctx := g.baseContext()
		ctx["Tag"] = tag
		ctx["Posts"] = posts

		outDir := filepath.Join(g.OutputDir, "tag", tag.Name)
		os.MkdirAll(outDir, 0755)
		if err := g.renderTemplate("tag.html", ctx, filepath.Join(outDir, "index.html")); err != nil {
			return err
		}
	}

	return nil
}

func (g *StaticGenerator) generatePages() error {
	log.Println("📃 生成独立页面...")

	pages, _ := ListPages()
	for _, page := range pages {
		content := RenderMarkdownPreview(page.Content)

		ctx := g.baseContext()
		ctx["Page"] = page
		ctx["Content"] = content

		outDir := filepath.Join(g.OutputDir, "page")
		os.MkdirAll(outDir, 0755)
		outPath := filepath.Join(outDir, page.Slug+".html")

		if err := g.renderTemplate("page.html", ctx, outPath); err != nil {
			return err
		}
	}

	return nil
}

func (g *StaticGenerator) generateRSS() error {
	log.Println("📡 生成 RSS...")

	posts, _ := GetPaginatedPosts(1, 20)

	var items strings.Builder
	for _, post := range posts {
		items.WriteString(fmt.Sprintf(`
    <item>
      <title>%s</title>
      <link>%s/%s/%s.html</link>
      <pubDate>%s</pubDate>
      <description><![CDATA[%s]]></description>
    </item>`,
			post.Title,
			g.BaseURL, post.Category, post.Slug,
			post.Date.Format("Mon, 02 Jan 2006 15:04:05 -0700"),
			post.Summary,
		))
	}

	rss := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>%s</title>
    <link>%s</link>
    <description>%s</description>
    <lastBuildDate>%s</lastBuildDate>
    %s
  </channel>
</rss>`,
		AppConfig.Site.Title,
		g.BaseURL,
		AppConfig.Site.Description,
		time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700"),
		items.String(),
	)

	return os.WriteFile(filepath.Join(g.OutputDir, "feed.xml"), []byte(rss), 0644)
}

func (g *StaticGenerator) generateSitemap() error {
	log.Println("🗺️ 生成 sitemap...")

	var urls strings.Builder
	urls.WriteString(fmt.Sprintf("  <url><loc>%s/</loc></url>\n", g.BaseURL))

	// 文章
	posts, _ := GetPaginatedPosts(1, 10000)
	for _, post := range posts {
		urls.WriteString(fmt.Sprintf("  <url><loc>%s/%s/%s.html</loc></url>\n",
			g.BaseURL, post.Category, post.Slug))
	}

	// 分类
	categories, _ := ListCategories()
	for _, cat := range categories {
		urls.WriteString(fmt.Sprintf("  <url><loc>%s/category/%s/</loc></url>\n",
			g.BaseURL, cat.Name))
	}

	// 标签
	for _, tag := range ListTags() {
		urls.WriteString(fmt.Sprintf("  <url><loc>%s/tag/%s/</loc></url>\n",
			g.BaseURL, tag.Name))
	}

	// 独立页面
	pages, _ := ListPages()
	for _, page := range pages {
		urls.WriteString(fmt.Sprintf("  <url><loc>%s/page/%s.html</loc></url>\n",
			g.BaseURL, page.Slug))
	}

	sitemap := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
%s</urlset>`, urls.String())

	return os.WriteFile(filepath.Join(g.OutputDir, "sitemap.xml"), []byte(sitemap), 0644)
}

func (g *StaticGenerator) baseContext() pongo2.Context {
	categories, _ := ListCategories()
	pages, _ := ListVisiblePages()
	
	return pongo2.Context{
		"Site":          AppConfig.Site,
		"NavCategories": categories,
		"NavPages":      pages,
	}
}

func (g *StaticGenerator) renderTemplate(name string, ctx pongo2.Context, outPath string) error {
	tpl, err := g.templates.FromFile(name)
	if err != nil {
		return fmt.Errorf("加载模板 %s 失败: %v", name, err)
	}

	out, err := tpl.Execute(ctx)
	if err != nil {
		return fmt.Errorf("渲染模板 %s 失败: %v", name, err)
	}

	return os.WriteFile(outPath, []byte(out), 0644)
}

// copyDir 复制目录
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		return copyFile(path, dstPath)
	})
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	os.MkdirAll(filepath.Dir(dst), 0755)
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
