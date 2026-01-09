package pkg

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Site          SiteConfig
	Theme         string
	PostsPerPage  int `mapstructure:"posts_per_page"`
	Search        SearchConfig
	AdminUsername string
	AdminPassword string
	JWTSecret     string
	Port          string
}

type SiteConfig struct {
	Title       string
	Description string
	BaseURL     string `mapstructure:"base_url"`
}

type SearchConfig struct {
	IndexPath string `mapstructure:"index_path"`
}

var AppConfig Config

// ContentBasePath 内容目录的基础路径
var ContentBasePath string

func InitConfig() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load config.yaml
	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	// Override with env vars
	AppConfig.AdminUsername = os.Getenv("ADMIN_USERNAME")
	AppConfig.AdminPassword = os.Getenv("ADMIN_PASSWORD")
	AppConfig.JWTSecret = os.Getenv("JWT_SECRET")
	AppConfig.Port = os.Getenv("PORT")
	if AppConfig.Port == "" {
		AppConfig.Port = "8080"
	}

	// 校验管理员凭据
	if AppConfig.AdminUsername == "" || AppConfig.AdminPassword == "" {
		log.Println("WARNING: ADMIN_USERNAME or ADMIN_PASSWORD not set, admin panel will be inaccessible")
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
