package parser

import (
	"reflect"
	"testing"
)

func TestParseRegex(t *testing.T) {
	content := "http://www.baidu.com"
	pattern0 := "http://www.(.+)"
	ret0, err := ParseRegex(content, pattern0)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(ret0, []string{"baidu.com"}) {
		t.Error("pattern:", pattern0, ", result:", ret0)
	}
	t.Log(ret0)
}

func TestUrlEncode(t *testing.T) {
	var urls = []string{
		"https://www.baidu.com/s?wd=4月1号保险新规",
		"alkasjdfkl",
		"this is cool",
		"!@#$^^&&"}
	for _, url := range urls {
		res, err := UrlEncode(url)
		if err != nil {
			t.Error(err)
		}
		t.Log(res)
	}
}

func TestMakeAbsoluteUrl(t *testing.T) {
	baseurl := "http://www.baidu.com/a/b/index.html"
	var testcases = [][]string{
		[]string{"../../abc.php", "http://www.baidu.com/abc.php"},
		[]string{"xyz.php", "http://www.baidu.com/a/b/xyz.php"},
		[]string{"/ok.jsp", "http://www.baidu.com/ok.jsp"},
		[]string{"../d/e/f/../1.do", "http://www.baidu.com/a/d/e/1.do"},
	}
	for _, two := range testcases {
		res, err := MakeAbsoluteUrl(two[0], baseurl)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(res, two[1]) {
			t.Error(res, "!=", two[1])
		}
		t.Log(baseurl, two[0], res)
	}
}
