package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var info *log.Logger

func init() {
	file, err := os.OpenFile("./log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}
	info = log.New(io.MultiWriter(file, os.Stderr), "INFO: ", log.Ldate|log.Ltime)
}

func Log(v ...interface{}) {
	info.Println(v)
}

type AwemeList struct {
	Desc string `json:"desc"`
	ShareUrl string `json:"share_url"`
}

type SerachReuslt struct {
	StatusCode int `json:"status_code"`
	AwemeList []AwemeList `json:"aweme_list"`
}

type DeviceResult struct {
	InstallId string `json:"install_id_str"`
	DeviceIdStr string `json:"device_id_str"`
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

func getDevice()  (DeviceResult, error){
	data, err := httpGet("http://192.168.1.12:8005/device_id")
	result := DeviceResult{}
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(data, &result)
	if err == nil {
		return result, err
	}
	return result, err
}

func main() {

	wordData := ReadFileData("./words.txt")
	device, err := getDevice()
	if err != nil {
		fmt.Println("获取设备号失败")
	}

	if wordData != "" {
		words := strings.Split(wordData, "\r\n")
		for _, word := range words {
			data, err := httpGet("http://192.168.1.12:8005/discover_search?device_id=" + device.DeviceIdStr + "&iid="  +device.InstallId + "&keyword=" + word)
			if err != nil {
				Log(err)
			}
			result := SerachReuslt{}
			err = json.Unmarshal(data, &result)
			if err == nil {
				if len(result.AwemeList) == 0 {
					Log("获取数据失败，更换设备号 关键词:" + word)
					device, err = getDevice()
					if err != nil {
						Log("获取设备号失败")
						break
					}
					Log("获取设备号成功")
					continue
				} else {
					Log("关键词:" + word)
					for _, value := range result.AwemeList {
						Log(value.Desc, value.ShareUrl + word)
					}
				}
			}
			time.Sleep(time.Second * 1)
		}
	}

}