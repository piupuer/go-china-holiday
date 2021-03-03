<h1 align="center">China Holiday</h1>

<div align="center">
基于colly实现的中国节假日查询的Go语言工具
<p align="center">
<img src="https://img.shields.io/github/go-mod/go-version/piupuer/go-china-holiday" alt="Go version"/>
<img src="https://img.shields.io/badge/Colly-v1.2.0-brightgreen" alt="Gin version"/>
<img src="https://img.shields.io/github/license/piupuer/go-china-holiday" alt="License"/>
</p>
</div>

## 示例

```go
package main

import (
  "fmt"
  holiday "github.com/piupuer/go-china-holiday"
)

func main() {
  // 初始化实例
  h, err := holiday.New(&holiday.Config{
    // 存储至本地文件, 减少每次从线上获取
    Filename: "holiday-data",
  })
  if err != nil {
    panic(err)
  }
  b, err := h.Check("2021-01-01")
  if err != nil {
    panic(err)
  }
  // 检查指定日期是否是法定节假日
  fmt.Printf("2021-01-01是否节假日: %v\n", b)
  
  // 列举2021年所有节假日
  holidays, workdays, err := h.List(2021)
  if err != nil {
    panic(err)
  }
  fmt.Println("2021年:\n", "节假日", holidays, "\n调休日（需要上班）", workdays)
  
  // 查询指定时间范围节假日(注意：将来的节假日需等gov.cn发布)
  holidays, workdays, err = h.Range("2020-04-01", "2021-12-01")
  if err != nil {
    panic(err)
  }
  fmt.Println("2020-04-01至2021-12-01:\n", "节假日", holidays, "\n调休日（需要上班）", workdays)
}
```

## 互动交流

### 与作者对话

> 加群一起学习一起进步

### QQ群：943724601

<img src="https://github.com/piupuer/gin-web-images/blob/master/contact/qq_group.jpeg?raw=true" width="256" alt="QQ群" />

## MIT License

    Copyright (c) 2021 piupuer
