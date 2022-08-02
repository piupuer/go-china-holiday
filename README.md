<h1 align="center">China Holiday</h1>

<div align="center">
基于colly实现的中国节假日查询的Go语言工具(只支持2008年及以后的节假日数据)
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
	h := holiday.New(holiday.WithFilename("holiday-data.txt"))
	// 检查指定日期是否是法定节假日
	fmt.Printf("2019-01-01是否节假日: %v\n\n", h.Check("2019-01-01"))

	// 列举2016年和2019所有节假日
	holidays, workdays := h.List(2016, 2019)

	fmt.Printf("2016年和2019年:\n节假日%v\n调休日（需要上班）%v\n\n", holidays, workdays)

	// 查询指定时间范围节假日(注意：将来的节假日需等gov.cn发布)
	holidays, workdays = h.Range("2008-01-01", "2022-12-01")

	fmt.Printf("2008-01-01至2022-12-01:\n节假日%v\n调休日（需要上班）%v\n", holidays, workdays)
}
```

## 互动交流

### 与作者对话

> 加群一起学习一起进步

### QQ群：943724601

<img src="https://github.com/piupuer/gin-web-images/blob/master/contact/qq_group.jpeg?raw=true" width="256" alt="QQ群" />

## MIT License

    Copyright (c) 2021 piupuer
