package parser

import (
	"golang.org/x/net/idna"
	"net/url"
	"regexp"
)

func ParseRegex(content, pattern string) ([]string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	var ret []string
	res := re.FindAllStringSubmatch(content, -1)
	for i, _ := range res {
		switch {
		case len(res[i]) == 1:
			ret = append(ret, res[i][0])
		case len(res[i]) > 1:
			ret = append(ret, res[i][1:]...)
		}
	}
	return ret, nil
}

func MatchRegex(content, pattern string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(content)
}

func UrlEncode(rawurl string) (string, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	u.Host, err = idna.ToASCII(u.Host)
	if err != nil {
		return "", err
	}
	u.RawQuery = u.Query().Encode()
	return u.String(), nil
}

func MakeAbsoluteUrl(href, baseurl string) (string, error) {
	u, err := url.Parse(href)
	if err != nil {
		return "", err
	}
	base, err := url.Parse(baseurl)
	if err != nil {
		return "", err
	}
	u = base.ResolveReference(u)
	return u.String(), nil
}
