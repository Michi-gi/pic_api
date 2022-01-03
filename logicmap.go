package main

var LogicMap = map[string]SiteAccess{"pixiv": PixivAccess{}}

type SiteId struct {
	Kind string `json:"kind"`
	Site string `json:"site"`
	Id   string `json:"id"`
}
