---
title: "03. 青峰 Swagger 主题配置：打造个性化文档界面"
date: 2026-01-10
tags: [青峰, Swagger, Go, 主题]
---

青峰 Swagger 提供丰富的主题配置选项，让你的 API 文档与众不同。

---

## 三种 UI 主题

### 1. Default - 经典主题

```go
qingfeng.Config{
    UITheme: qingfeng.ThemeDefault,
}
```

特点：
- 经典蓝色风格
- 功能完整
- 适合大多数场景

### 2. Minimal - 简约主题

```go
qingfeng.Config{
    UITheme: qingfeng.ThemeMinimal,
}
```

特点：
- 黑白极简设计
- 专业干净
- 适合正式文档

### 3. Modern - 现代主题

```go
qingfeng.Config{
    UITheme: qingfeng.ThemeModern,
}
```

特点：
- 渐变毛玻璃效果
- 视觉冲击力强
- 适合产品展示

## 深色/浅色模式

### 代码配置

```go
qingfeng.Config{
    DarkMode: true,  // 默认深色模式
}
```

### URL 参数切换

```
http://localhost:8080/doc/?dark=true
http://localhost:8080/doc/?dark=false
```

### 界面切换

点击页面右上角的 🌙/☀️ 图标即可切换。

## 六种主题色

青峰支持 6 种主题色：

| 颜色 | 说明 |
|------|------|
| 蓝色 | 默认，专业稳重 |
| 绿色 | 清新自然 |
| 紫色 | 优雅神秘 |
| 橙色 | 活力热情 |
| 红色 | 醒目强烈 |
| 青色 | 清爽现代 |

在界面中点击「主题」按钮可以切换主题色。

## 自定义 Logo

### 使用 URL

```go
qingfeng.Config{
    Logo:     "https://example.com/logo.png",
    LogoLink: "https://example.com",  // 点击跳转
}
```

### 使用 Base64

```go
qingfeng.Config{
    Logo: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...",
}
```

### Logo 尺寸建议

- 高度：32-48px
- 格式：PNG/SVG（支持透明背景）
- 宽度：自动按比例缩放

## 完整配置示例

```go
r.GET("/doc/*any", qingfeng.Handler(qingfeng.Config{
    // 基础信息
    Title:       "我的 API",
    Description: "这是一个示例 API 文档",
    Version:     "2.0.0",
    BasePath:    "/doc",
    DocPath:     "./docs/swagger.json",
    
    // 主题配置
    UITheme:  qingfeng.ThemeModern,
    DarkMode: false,
    
    // 自定义 Logo
    Logo:     "https://example.com/logo.png",
    LogoLink: "https://example.com",
}))
```

## URL 参数覆盖

可以通过 URL 参数临时覆盖配置：

```
# 切换主题
/doc/?theme=modern
/doc/?theme=minimal
/doc/?theme=default

# 切换深色模式
/doc/?dark=true
```

## 设置持久化

用户的主题设置会自动保存到浏览器本地存储：

- 主题风格
- 深色/浅色模式
- 主题色
- 全局参数

下次访问时会自动恢复上次的设置。

---

**下一篇**: [在线调试 API](/qingfeng/04-debugging.html)
