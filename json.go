package main

type Data struct {
	AwemeList  []Item `json:"item_list"`
}

type Item struct {
	AwemeId      string      `json:"aweme_id"`
	Desc         string      `json:"desc"`
	Video        video       `json:"video"`
}

type video struct {
	DownloadAddr uriStr `json:"download_addr"`
	PlayAddr     uriStr `json:"play_addr"`
	OriginCover  uriStr `json:"origin_cover"`
	Cover        uriStr `json:"cover"`
	Vid          string `json:"vid"`
}

type uriStr struct {
	Uri     string   `json:"uri"`
	UrlList []string `json:"url_list"`
	width   int      `json:"width"`
	height  int      `json:"height"`
}
