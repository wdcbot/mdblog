package theme

import (
	"html/template"
	"mdblog/internal/pkg"
	"net/http"
	"path/filepath"
	"sort"

	"github.com/flosch/pongo2/v6"
	"github.com/gin-gonic/gin"
)

var (
	ThemeSet       *pongo2.TemplateSet
	AdminTemplates *template.Template // 保持后台使用原生模板
)

func InitPongo2() {
	themeName := pkg.AppConfig.Theme
	if themeName == "" {
		themeName = "pure"
	}
	
	// 设置 Pongo2 模板加载路径 (前台)
	loader := pongo2.MustNewLocalFileSystemLoader(filepath.Join("themes", themeName, "layouts"))
	ThemeSet = pongo2.NewSet(themeName, loader)
}

// LoadAdminTemplates 仅加载后台模板
func LoadAdminTemplates() {
	funcMap := template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}

	adminDir := filepath.Join("admin", "layouts")
	// 注意：这里只解析 admin/layouts 下的模板，不再解析 themes 目录
	AdminTemplates = template.Must(template.New("admin").Funcs(funcMap).ParseGlob(filepath.Join(adminDir, "*.html")))
}

// Render Pongo2 统一渲染函数 (用于前台)
func Render(c *gin.Context, templateName string, data gin.H) {
	tmpl, err := ThemeSet.FromFile(templateName)
	if err != nil {
		c.String(http.StatusInternalServerError, "Template Error: "+err.Error())
		return
	}

	cats, _ := pkg.ListCategories()
	sort.Slice(cats, func(i, j int) bool {
		return cats[i].Name < cats[j].Name
	})

	pages, _ := pkg.ListVisiblePages()

	ctx := pongo2.Context{
		"Site":          pkg.AppConfig.Site,
		"NavCategories": cats,
		"NavPages":      pages,
	}
	for k, v := range data {
		ctx[k] = v
	}

	err = tmpl.ExecuteWriter(ctx, c.Writer)
	if err != nil {
		c.String(http.StatusInternalServerError, "Render Error: "+err.Error())
	}
}
