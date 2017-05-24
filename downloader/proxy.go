package downloader

import (
	"errors"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	proxyHost   = "http://127.0.0.1:8888"
	proxyGet    = "/get"
	proxyReport = "/report"
)

func GetProxy() (string, error) {
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	url := proxyHost + proxyGet
	for i := 0; i < 3; i++ {
		resp, err := client.Get(url)
		if err != nil {
			glog.Error("failed to get proxy, retry: ", i, "msg: ", err)
			continue
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			glog.Error("failed to read proxy, retry: ", i, "msg: ", err)
			continue
		}
		return "http://" + string(body), nil
	}
	return "", errors.New("failed to get proxy finally")
}

func ReportProxyStatus(proxy string, isValid bool) error {
	if len(proxy) == 0 {
		return nil
	}
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	reportUrl := proxyHost + proxyReport
	data := url.Values{}
	data.Add("proxy", proxy)
	if isValid {
		data.Add("status", "valid")
	} else {
		data.Add("status", "invalid")
	}
	resp, err := client.Post(reportUrl, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
