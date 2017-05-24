package downloader

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/axgle/mahonia"
	"github.com/crawlerclub/x/types"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func Download(requestInfo *types.HttpRequest) *types.HttpResponse {
	var timeout time.Duration
	if requestInfo.Timeout > 0 {
		timeout = time.Duration(requestInfo.Timeout) * time.Second
	} else {
		timeout = 30 * time.Second
	}
	client := &http.Client{
		Timeout: timeout,
	}
	responseInfo := &types.HttpResponse{
		Url: requestInfo.Url,
	}
	transport := http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
	}

	//proxy
	if requestInfo.UseProxy {
		var proxy string
		var err error
		if len(requestInfo.Proxy) > 0 {
			proxy = requestInfo.Proxy
		} else {
			proxy, err = GetProxy()
			if err != nil {
				responseInfo.Error = err
				return responseInfo
			}
		}
		responseInfo.Proxy = proxy
		urlProxy, err := url.Parse(proxy)
		if err != nil {
			responseInfo.Error = errors.New(fmt.Sprintf("failed to parse proxy: %s", proxy))
			return responseInfo
		}
		transport.Proxy = http.ProxyURL(urlProxy)
	}

	client.Transport = &transport

	req, err := http.NewRequest(requestInfo.Method, requestInfo.Url, strings.NewReader(requestInfo.PostData))
	if err != nil {
		responseInfo.Error = err
		return responseInfo
	}
	var resp *http.Response
	resp, err = client.Do(req)
	if err != nil {
		responseInfo.Error = err
		return responseInfo
	}

	responseInfo.StatusCode = resp.StatusCode
	defer resp.Body.Close()

	var contentLen int64
	contentLen, err = strconv.ParseInt(resp.Header.Get("content-length"), 10, 64)
	if err != nil {
		//
	} else if requestInfo.MaxLen > 0 && contentLen > requestInfo.MaxLen {
		responseInfo.Error = errors.New("reponse size too large")
		return responseInfo
	}

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		if reader, err = gzip.NewReader(resp.Body); err != nil {
			responseInfo.Error = err
			return responseInfo
		}
		defer reader.Close()
	case "deflate":
		if reader, err = zlib.NewReader(resp.Body); err != nil {
			responseInfo.Error = err
			return responseInfo
		}
		defer reader.Close()
	default:
		reader = resp.Body
	}

	var readLen int64 = 0
	respBuf := bytes.NewBuffer([]byte{})
	for {
		readData := make([]byte, 4096)
		length, err := reader.Read(readData)
		respBuf.Write(readData[:length])
		readLen += int64(length)
		if err != nil {
			if err == io.EOF {
				break
			}
			responseInfo.Error = errors.New("reponse size too large - count")
			return responseInfo
		}
	}
	responseInfo.Content = respBuf.Bytes()
	var encoding string
	encoding, err = GuessEncoding(responseInfo.Content)
	if err != nil {
		//
		responseInfo.Text = string(responseInfo.Content)
		responseInfo.Encoding = ""
		return responseInfo
	}
	encoder := mahonia.NewDecoder(encoding)
	if encoder == nil {
		responseInfo.Text = string(responseInfo.Content)
		responseInfo.Encoding = ""
		return responseInfo
	}
	responseInfo.Text = encoder.ConvertString(string(responseInfo.Content))
	responseInfo.Encoding = encoding
	return responseInfo
}
