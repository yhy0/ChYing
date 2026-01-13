package qqwry

// https://github.com/zu1k/nali

import (
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/yhy0/logging"
)

const (
	// 使用 metowolf 的 GitHub 镜像，直接提供解密后的 qqwry.dat
	qqwryURL = "https://github.com/metowolf/qqwry.dat/releases/latest/download/qqwry.dat"
)

func Download(filePath string) error {
	data, err := downloadQQwry()
	if err != nil {
		return err
	}
	return SaveFile(filePath, data)
}

func downloadQQwry() ([]byte, error) {
	return Get(qqwryURL)
}

func Get(url string) ([]byte, error) {
	var UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36"
	Client := &http.Client{
		Timeout: time.Second * 60, // 增加超时时间，因为文件较大
		Transport: &http.Transport{
			TLSHandshakeTimeout:   time.Second * 10,
			IdleConnTimeout:       time.Second * 30,
			ResponseHeaderTimeout: time.Second * 30,
			ExpectContinueTimeout: time.Second * 30,
		},
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("User-Agent", UserAgent)
	resp, err := Client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("http response is nil")
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("http response status code is not 200")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func SaveFile(path string, data []byte) (err error) {
	// Remove file if exist
	_, err = os.Stat(path)
	if err == nil {
		err = os.Remove(path)
		if err != nil {
			logging.Logger.Errorln("旧文件删除失败" + err.Error())
		}
	}

	// save file
	return os.WriteFile(path, data, 0644)
}
