---
title: "RESTful API 设计原则"
date: 2026-01-09
tags: [API, REST, 后端]
---

好的 API 设计能让开发者更容易理解和使用你的服务。

## 核心原则

### 1. 使用名词而非动词

```
✅ GET /users
❌ GET /getUsers
```

### 2. 使用复数形式

```
✅ /users
❌ /user
```

### 3. 合理使用 HTTP 方法

| 方法 | 用途 |
|------|------|
| GET | 获取资源 |
| POST | 创建资源 |
| PUT | 更新资源（全量） |
| PATCH | 更新资源（部分） |
| DELETE | 删除资源 |

## 状态码

```
200 OK - 成功
201 Created - 创建成功
400 Bad Request - 请求参数错误
401 Unauthorized - 未认证
404 Not Found - 资源不存在
500 Internal Server Error - 服务器错误
```

## 版本控制

推荐在 URL 中包含版本号：

```
/api/v1/users
/api/v2/users
```

好的 API 是产品的门面。
