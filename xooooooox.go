// Copyright (C) xooooooox
// datetime 2019-12-26 11:04:11

package main

// Resources resources 资源表
type Resources struct {
	Id      int64  `json:"id"`      // ID
	Type    int64  `json:"type"`    // 类型
	Group   int64  `json:"group"`   // 群组
	Keyword string `json:"keyword"` // 关键字
	Title   string `json:"title"`   // 标题
	Uri     string `json:"uri"`     // 地址
	Owner   int64  `json:"owner"`   // 拥有者
	Status  int64  `json:"status"`  // 状态
	Time    int64  `json:"time"`    // 时间
	Note    string `json:"note"`    // 备注
}
