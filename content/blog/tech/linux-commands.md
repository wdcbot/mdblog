---
title: "Linux 常用命令速查"
date: 2025-12-20
tags: [Linux, 命令行, 运维]
---

整理一些日常开发中常用的 Linux 命令。

## 文件操作

```bash
# 查看文件内容
cat file.txt
less file.txt
tail -f log.txt  # 实时查看日志

# 查找文件
find /path -name "*.go"
locate filename

# 文件权限
chmod 755 script.sh
chown user:group file
```

## 进程管理

```bash
# 查看进程
ps aux | grep nginx
top
htop

# 杀死进程
kill -9 <pid>
pkill nginx
```

## 网络相关

```bash
# 端口查看
netstat -tlnp
ss -tlnp
lsof -i :8080

# 网络请求
curl -X GET http://api.example.com
wget http://example.com/file.zip
```

## 磁盘空间

```bash
df -h      # 查看磁盘使用
du -sh *   # 查看目录大小
```

熟练使用命令行能大大提升效率。
