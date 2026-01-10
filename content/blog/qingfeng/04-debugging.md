---
title: "04. 青峰 Swagger 在线调试：告别 Postman"
date: 2026-01-10
tags: [青峰, Swagger, Go, 调试]
---

青峰 Swagger 内置强大的 API 调试工具，让你无需离开文档页面就能测试接口。

---

## 启用调试功能

调试功能默认开启，也可以手动配置：

```go
qingfeng.Config{
    EnableDebug: true,  // 默认 true
}
```

## 基本使用

1. 点击任意接口展开详情
2. 点击「调试」按钮打开调试面板
3. 填写参数
4. 点击「发送请求」
5. 查看响应结果

## 全局请求头

### 代码预设

```go
qingfeng.Config{
    GlobalHeaders: []qingfeng.Header{
        {Key: "Authorization", Value: "Bearer your-token"},
        {Key: "X-API-Key", Value: "your-api-key"},
    },
}
```

### 界面配置

1. 点击页面顶部的「全局参数」按钮
2. 添加 Header 名称和值
3. 保存后所有请求都会自动带上这些 Header

## Token 自动提取

登录接口返回 Token 后，可以自动提取并设置到全局参数：

1. 点击「Token」按钮
2. 配置提取规则：
   - **响应字段路径**: `data.token` 或 `data.access_token`
   - **目标 Header**: `Authorization`
   - **前缀**: `Bearer `（注意空格）
3. 调用登录接口后，Token 会自动设置

### 示例

假设登录接口返回：
```json
{
  "code": 0,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

配置：
- 响应字段路径: `data.token`
- 目标 Header: `Authorization`
- 前缀: `Bearer `

调用登录后，后续请求会自动带上：
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

## 多环境切换

配置多个环境：

```go
qingfeng.Config{
    Environments: []qingfeng.Environment{
        {Name: "本地开发", BaseURL: "http://localhost:8080/api"},
        {Name: "测试环境", BaseURL: "https://test-api.example.com/api"},
        {Name: "生产环境", BaseURL: "https://api.example.com/api"},
    },
}
```

在页面顶部选择环境，请求会发送到对应的 BaseURL。

## 请求体模板

对于复杂的请求体，可以保存为模板：

1. 在请求体输入框填写 JSON
2. 点击「保存模板」
3. 输入模板名称（如「创建普通用户」「创建管理员」）
4. 下次点击「模板」按钮选择已保存的模板

模板按接口保存，每个接口可以有多个模板。

## 参数启用/禁用

每个参数前有勾选框：

- ✅ 勾选：参数会发送
- ⬜ 不勾选：参数不发送（显示半透明）

适用于测试可选参数的场景。

## 复制 cURL

点击「复制 cURL」按钮，可以复制当前请求的 cURL 命令：

```bash
curl -X POST 'http://localhost:8080/api/users' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer xxx' \
  -d '{"name":"张三","email":"test@example.com"}'
```

方便在终端中调试或分享给同事。

## 快捷键

| 快捷键 | 功能 |
|--------|------|
| `Ctrl/Cmd + K` | 聚焦搜索框 |
| `Ctrl/Cmd + Enter` | 发送请求 |
| `Escape` | 关闭弹窗 |

## 响应结果

响应结果支持：

- **JSON 语法高亮**: 彩色显示，易于阅读
- **格式化**: 自动缩进
- **复制**: 一键复制响应内容
- **状态码**: 显示 HTTP 状态码和耗时

## 文件上传

对于文件上传接口：

1. 参数类型会显示为「文件」
2. 点击「选择文件」按钮
3. 选择本地文件
4. 发送请求

## 调试数据持久化

调试时填写的参数会自动保存到 `sessionStorage`：

- 刷新页面后参数不丢失
- 关闭浏览器后清除
- 可通过配置禁用：

```go
persist := false
qingfeng.Config{
    PersistParams: &persist,
}
```

---

**下一篇**: [生产环境部署](/qingfeng/05-deployment.html)
