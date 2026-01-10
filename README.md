# mdblog

轻量级 Markdown 博客系统，无数据库依赖。

> 本项目基于 [TwoThreeWang/mdblog](https://github.com/TwoThreeWang/mdblog) 二次开发，进行了大量功能增强和优化。

**预览地址：** https://wdc.zeabur.app

## 分支说明

- `main` - 包含示例文章和文档，用于预览和学习
- `template` - 干净的模板分支，**个人使用请 Fork 此分支**

## 特性

- 📄 Markdown 文件存储
- ⚡ 内存缓存，高性能
- 🔍 内置全文搜索
- 🌙 暗色模式
- 💬 评论系统
- 📱 响应式设计
- 🎨 后台可视化编辑

## 快速部署到 Zeabur

1. Fork 这个仓库的 `template` 分支到你的 GitHub
2. 去 [Zeabur](https://zeabur.com) 用 GitHub 登录
3. 创建项目 → 选择「共享集群」→ 选择「香港」地区
4. 添加服务 → Git → 选择你 fork 的仓库
5. 等待构建完成，生成域名即可访问

## 本地运行

```bash
# 使用干净的 template 分支
git clone -b template https://github.com/wdcbot/mdblog.git
cd mdblog
go run main.go
```

- 前台：http://localhost:8080
- 后台：http://localhost:8080/admin（默认 admin / admin888）

## 文档

详细使用说明请访问：https://wdc.zeabur.app

## License

MIT
