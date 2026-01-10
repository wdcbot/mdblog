package pkg

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Site          SiteConfig
	Theme         string
	PostsPerPage  int `mapstructure:"posts_per_page"`
	Search        SearchConfig
	Server        ServerConfig
	Admin         AdminConfig
	AdminUsername string
	AdminPassword string
	JWTSecret     string
	Port          string
}

type ServerConfig struct {
	Port string
}

type AdminConfig struct {
	Username  string
	Password  string
	JWTSecret string `mapstructure:"jwt_secret"`
}

type SiteConfig struct {
	Title              string
	Description        string
	Keywords           string
	Author             string
	BaseURL            string `mapstructure:"base_url"`
	Favicon            string
	Logo               string
	LogoHeight         int    `mapstructure:"logo_height"` // Logo 高度（像素）
	HeroTitle          string `mapstructure:"hero_title"`
	CommentsEnabled    bool   `mapstructure:"comments_enabled"`
	TOCEnabled         bool   `mapstructure:"toc_enabled"`
	ReadingTimeEnabled bool   `mapstructure:"reading_time_enabled"`
	DefaultTheme       string `mapstructure:"default_theme"`
	AccentColor        string `mapstructure:"accent_color"`
	Analytics          string
	AdsCode            string `mapstructure:"ads_code"` // 广告代码（如 Google AdSense）
	Footer             FooterConfig
}

type FooterConfig struct {
	Copyright string
	ICP       string
	Links     []FooterLink
}

type FooterLink struct {
	Name string
	URL  string
	Icon string
}

type SearchConfig struct {
	IndexPath string `mapstructure:"index_path"`
}

var AppConfig Config

// ContentBasePath 内容目录的基础路径
var ContentBasePath string

func InitConfig() {
	// Load config.yaml
	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	// 从 config.yaml 读取配置
	AppConfig.AdminUsername = AppConfig.Admin.Username
	AppConfig.AdminPassword = AppConfig.Admin.Password
	AppConfig.JWTSecret = AppConfig.Admin.JWTSecret
	AppConfig.Port = AppConfig.Server.Port
	
	// 环境变量可以覆盖配置文件（用于 Docker 等场景）
	if envPort := os.Getenv("PORT"); envPort != "" {
		AppConfig.Port = envPort
	}
	if envUser := os.Getenv("ADMIN_USERNAME"); envUser != "" {
		AppConfig.AdminUsername = envUser
	}
	if envPass := os.Getenv("ADMIN_PASSWORD"); envPass != "" {
		AppConfig.AdminPassword = envPass
	}
	if envSecret := os.Getenv("JWT_SECRET"); envSecret != "" {
		AppConfig.JWTSecret = envSecret
	}
	
	// 默认端口
	if AppConfig.Port == "" {
		AppConfig.Port = "8080"
	}

	// 校验管理员凭据
	if AppConfig.AdminUsername == "" || AppConfig.AdminPassword == "" {
		log.Println("WARNING: admin username or password not set, admin panel will be inaccessible")
	}

	// 初始化内容基础路径（用于路径安全校验）
	ContentBasePath, _ = filepath.Abs("content")
}

// IsPathSafe 检查路径是否在允许的目录内，防止目录遍历攻击
func IsPathSafe(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	return strings.HasPrefix(absPath, ContentBasePath)
}

func UpdateConfig(siteTitle, siteDesc, baseURL, theme string, postsPerPage int) error {
	// Update in-memory config
	AppConfig.Site.Title = siteTitle
	AppConfig.Site.Description = siteDesc
	AppConfig.Site.BaseURL = baseURL
	AppConfig.Theme = theme
	AppConfig.PostsPerPage = postsPerPage

	// Update viper registry
	viper.Set("site.title", siteTitle)
	viper.Set("site.description", siteDesc)
	viper.Set("site.base_url", baseURL)
	viper.Set("theme", theme)
	viper.Set("posts_per_page", postsPerPage)

	// Write to file
	return viper.WriteConfig()
}

// UpdateBasicConfig 更新基础配置
func UpdateBasicConfig(siteTitle, siteDesc, baseURL, theme string, postsPerPage int, keywords, author string) error {
	AppConfig.Site.Title = siteTitle
	AppConfig.Site.Description = siteDesc
	AppConfig.Site.BaseURL = baseURL
	AppConfig.Theme = theme
	AppConfig.PostsPerPage = postsPerPage
	AppConfig.Site.Keywords = keywords
	AppConfig.Site.Author = author

	viper.Set("site.title", siteTitle)
	viper.Set("site.description", siteDesc)
	viper.Set("site.base_url", baseURL)
	viper.Set("theme", theme)
	viper.Set("posts_per_page", postsPerPage)
	viper.Set("site.keywords", keywords)
	viper.Set("site.author", author)

	return viper.WriteConfig()
}

// UpdateAppearanceConfig 更新外观配置
func UpdateAppearanceConfig(heroTitle, defaultTheme, accentColor, favicon, logo string, logoHeight int) error {
	AppConfig.Site.HeroTitle = heroTitle
	AppConfig.Site.DefaultTheme = defaultTheme
	AppConfig.Site.AccentColor = accentColor
	AppConfig.Site.Favicon = favicon
	AppConfig.Site.Logo = logo
	AppConfig.Site.LogoHeight = logoHeight

	viper.Set("site.hero_title", heroTitle)
	viper.Set("site.default_theme", defaultTheme)
	viper.Set("site.accent_color", accentColor)
	viper.Set("site.favicon", favicon)
	viper.Set("site.logo", logo)
	viper.Set("site.logo_height", logoHeight)

	return viper.WriteConfig()
}

// UpdateFeaturesConfig 更新功能开关配置
func UpdateFeaturesConfig(commentsEnabled, tocEnabled, readingTimeEnabled bool) error {
	AppConfig.Site.CommentsEnabled = commentsEnabled
	AppConfig.Site.TOCEnabled = tocEnabled
	AppConfig.Site.ReadingTimeEnabled = readingTimeEnabled

	viper.Set("site.comments_enabled", commentsEnabled)
	viper.Set("site.toc_enabled", tocEnabled)
	viper.Set("site.reading_time_enabled", readingTimeEnabled)

	return viper.WriteConfig()
}

// UpdateFooterConfig 更新页脚配置
func UpdateFooterConfig(copyright, icp string, links []FooterLink) error {
	AppConfig.Site.Footer.Copyright = copyright
	AppConfig.Site.Footer.ICP = icp
	AppConfig.Site.Footer.Links = links

	viper.Set("site.footer.copyright", copyright)
	viper.Set("site.footer.icp", icp)
	
	// 转换 links 为 viper 可用的格式
	var linksData []map[string]string
	for _, link := range links {
		linksData = append(linksData, map[string]string{
			"name": link.Name,
			"url":  link.URL,
			"icon": link.Icon,
		})
	}
	viper.Set("site.footer.links", linksData)

	return viper.WriteConfig()
}

// UpdateAnalyticsConfig 更新统计代码配置
func UpdateAnalyticsConfig(analytics string) error {
	AppConfig.Site.Analytics = analytics
	viper.Set("site.analytics", analytics)
	return viper.WriteConfig()
}

// UpdateAdsConfig 更新广告代码配置
func UpdateAdsConfig(adsCode string) error {
	AppConfig.Site.AdsCode = adsCode
	viper.Set("site.ads_code", adsCode)
	return viper.WriteConfig()
}

func ListThemes() ([]string, error) {
	entries, err := os.ReadDir("themes")
	if err != nil {
		return nil, err
	}
	var themes []string
	for _, e := range entries {
		if e.IsDir() {
			themes = append(themes, e.Name())
		}
	}
	return themes, nil
}
