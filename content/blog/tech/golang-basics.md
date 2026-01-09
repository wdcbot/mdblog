---
title: "Go 语言入门指南"
date: 2026-01-08
tags: [Go, 编程, 后端]
---

Go 语言是 Google 开发的一种静态强类型、编译型语言。它具有简洁、高效、并发支持好等特点。

## 为什么选择 Go

1. **简洁的语法** - 没有类继承，没有泛型（1.18 之前），代码清晰易读
2. **高性能** - 编译为原生机器码，执行效率接近 C
3. **内置并发** - goroutine 和 channel 让并发编程变得简单

## Hello World

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

## 变量声明

Go 有多种变量声明方式：

```go
var name string = "Alice"
var age = 25
count := 100
```

推荐使用 `:=` 短声明，简洁高效。
