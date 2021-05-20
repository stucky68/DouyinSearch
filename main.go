package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	jsoniter "github.com/json-iterator/go"
)

var info *log.Logger

func init() {
	file, err := os.OpenFile("./log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}
	info = log.New(io.MultiWriter(file, os.Stderr), "INFO: ", log.Ldate|log.Ltime)
}

func GetData(url string) (itemID, dytk string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	rand.Seed(time.Now().Unix())
	s := strconv.Itoa(rand.Intn(1000))

	req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/" + s  + ".36 (KHTML, like Gecko) Chrome/75.0.3770.100 Mobile Safari/537.36")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("accept-language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Add("Cookie", "_ga=GA1.2.685263550.1587277283; _gid=GA1.2.143250871.1587911549; tt_webid=6820028204934923790; _ba=BA0.2-20200301-5199e-c7q9NP0laGm7KfaPfGcH")
	res, err := client.Do(req)
	if err == nil {
		b, _ := ioutil.ReadAll(res.Body)
		result := string(b)
		var itemIDRegexp = regexp.MustCompile(`itemId: "(.*?)"`)
		ids := itemIDRegexp.FindStringSubmatch(result)
		if len(ids) > 1 {
			itemID = ids[1]
		}
		var dytkRegexp = regexp.MustCompile(`dytk: "(.*?)"`)
		dytks := dytkRegexp.FindStringSubmatch(result)
		if len(dytks) > 1 {
			dytk = dytks[1]
		}
	}
	return
}

func read3(path string) string {
	fi, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	fd, err := ioutil.ReadAll(fi)
	return string(fd)
}

func Download(getUrl, saveFile string) error {
	client := &http.Client{}
	rand.Seed(time.Now().Unix())
	s := strconv.Itoa(rand.Intn(1000))
	req, err := http.NewRequest("GET", getUrl, nil)
	if err != nil {
		return err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 12_0 like Mac OS X) AppleWebKit/"+s+".1.38 (KHTML, like Gecko) Version/11.0 Mobile/15A372 Safari/"+s+".1")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("upgrade-insecure-requests", "1")
	req.Header.Add("pragma", "no-cache")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.Header.Get("content-length") == "0" {
		return errors.New("ip")
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(saveFile, b, 0666)
	if err != nil {
		return err
	}
	return nil
}

func downloadHttpFile(videoUrl string, localVideo string) error {
	timeout := 0
	for {
		err := Download(videoUrl, localVideo)
		if err == nil {
			break
		}
		Log(err)
		timeout++
		if timeout > 3 {
			return errors.New("超时三次失败")
		}
	}
	return nil
}

func FilterEmoji(content string) string {
	newContent := ""
	for _, value := range content {
		if unicode.Is(unicode.Han, value) || unicode.IsLetter(value) || unicode.IsDigit(value) || unicode.IsSpace(value) {
			newContent += string(value)
		}
	}
	return newContent
}

func IsExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}

func HandleJson(word string, data Data) bool {
	if len(data.AwemeList) == 0 {
		return false
	}
	item := data.AwemeList[0]

	item.Desc = strings.ReplaceAll(item.Desc, ":", "")
	item.Desc = strings.ReplaceAll(item.Desc, "?", "")
	item.Desc = strings.ReplaceAll(item.Desc, "\\", "")
	item.Desc = strings.ReplaceAll(item.Desc, "/", "")
	item.Desc = strings.ReplaceAll(item.Desc, "\"", "")
	item.Desc = strings.ReplaceAll(item.Desc, "*", "")
	item.Desc = strings.ReplaceAll(item.Desc, "<", "")
	item.Desc = strings.ReplaceAll(item.Desc, ">", "")
	item.Desc = strings.ReplaceAll(item.Desc, "|", "")
	item.Desc = strings.ReplaceAll(item.Desc, "\r", "")
	item.Desc = strings.ReplaceAll(item.Desc, "\n", "")
	item.Desc = FilterEmoji(item.Desc)

	localVideo := "download/" + word + "/" + item.Desc + item.AwemeId[0:13] + ".mp4"

	if IsExist(localVideo) == false {
		Log("开始处理数据:", item.Desc)
		//fmt.Println(item.Video.PlayAddr.UrlList[0])
		err := downloadHttpFile("https://aweme.snssdk.com/aweme/v1/play/?video_id="+item.Video.Vid+"&media_type=4&vr_type=0&improve_bitrate=0&is_play_url=1&is_support_h265=0&source=PackSourceEnum_PUBLISH", localVideo)
		//err := downloadHttpFile(item.Video.PlayAddr.UrlList[0], localVideo)

		if len(item.Video.OriginCover.UrlList) > 0 {
			err := downloadHttpFile(item.Video.OriginCover.UrlList[0], "download/"+word+"/"+item.AwemeId[0:13]+".jpg")
			if err != nil {
				Log("下载封面失败:", err)
			}
		} else if len(item.Video.Cover.UrlList) > 0 {
			err := downloadHttpFile(item.Video.Cover.UrlList[0], "download/"+word+"/"+item.AwemeId[0:13]+".jpg")
			if err != nil {
				Log("下载封面失败:", err)
			}
		}

		if err != nil {
			Log("下载视频失败:", err)
		} else {
			Log("下载视频成功:", localVideo)
			return true
		}
	} else {
		Log(item.Desc + " " + localVideo + "文件已存在，跳过")
	}
	return false
}

func GetVideo(itemID string) (error, Data) {
	client := &http.Client{}
	url := "https://www.iesdouyin.com/web/api/v2/aweme/iteminfo/?item_ids=" + itemID

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err, Data{}
	}
	rand.Seed(time.Now().Unix())
	s := strconv.Itoa(rand.Intn(1000))

	req.Header.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3")
	//req.Header.Add("accept-encoding", "gzip, deflate, br")
	req.Header.Add("accept-language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Add("cache-control", "max-age=0")
	req.Header.Add("cookie", "_ga=GA1.2.938284732.1578806304; _gid=GA1.2.1428838914.1578806305")
	req.Header.Add("upgrade-insecure-requests", "1")
	req.Header.Add("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/" + s + ".36 (KHTML, like Gecko) Chrome/75.0.3770.100 Mobile Safari/537.36")
	req.Header.Add("Host", "www.iesdouyin.com")
	res, err := client.Do(req)
	if err != nil {
		return err, Data{}
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err, Data{}
	}
	var data Data
	err = json.Unmarshal(b, &data)
	if err != nil {
		//fmt.Println(err, string(b))
		return err, Data{}
	}
	return nil, data
}

func Log(v ...interface{}) {
	info.Println(v)
}

type AwemeList struct {
	Desc string `json:"desc"`
	FromGroupId string `json:"fromGroupId"`
}

type SerachReuslt struct {
	ErrorCode int `json:"error_code"`
	StatusCode int `json:"status_code"`
	AwemeList []AwemeList `json:"awemeList"`
	HasMore bool `json:"hasMore"`
	Cursor int `json:"cursor"`
}

func httpGet(url string) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func ReadFileData(path string) string {
	fi, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	fd, err := ioutil.ReadAll(fi)
	return string(fd)
}

func main() {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	wordData := ReadFileData("./words.txt")

	fmt.Println("请输入需要采集的条数:")
	serachCount := 0
	_, err := fmt.Scanf("%d\r\n", &serachCount)
	if err != nil {
		panic(err)
	}

	if wordData != "" {
		words := strings.Split(wordData, "\r\n")
		for _, word := range words {
			cursor := "0"
			hasMore := true
			sucCount := 0
			Log("关键词:" + word + "开始下载")

			os.MkdirAll("download/"+word, os.ModePerm)

			for hasMore {
				escape := url.QueryEscape(word)
				data, err := httpGet("http://47.108.87.251:8081/douyin/crawler/searchVideo?token=5617ee22e807e0cc8467c6202dbaac02&pageNum=" + cursor + "&keyword=" + escape)
				if err != nil {
					Log(err)
					time.Sleep(time.Second * 10)
					continue
				}
				result := SerachReuslt{}

				fmt.Println(string(data))
				if len(data) == 0 {
					Log("获取数据数据为空 等待10秒后继续 关键词:" + word)
					time.Sleep(time.Second * 10)
					continue
				}

				err = json.Unmarshal(data, &result)
				if err == nil {
					if len(result.AwemeList) == 0 {
						Log("获取数据失败 关键词:" + word)
						time.Sleep(time.Second * 10)
						continue
					} else {
						for _, value := range result.AwemeList {
							if value.FromGroupId != "" {
								err, d := GetVideo(value.FromGroupId)
								if err == nil {
									if HandleJson(word, d) {
										sucCount++
									}
								}
							} else {
								Log(value.Desc + " 获取数据失败")
							}

							Log("下载成功第"+strconv.Itoa(sucCount)+"个视频", value.Desc, value.FromGroupId)

							if sucCount >= serachCount {
								break
							}
						}

						if sucCount >= serachCount {
							break
						}

						hasMore = result.HasMore
						cursor = strconv.Itoa(result.Cursor)
					}
				} else {
					Log("错误:" + err.Error())
					time.Sleep(time.Second * 10)
					continue
				}
				time.Sleep(time.Second * 1)
			}

			Log("关键词:" + word + "下载完毕")
		}
	}
}