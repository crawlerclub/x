package store

import (
	"encoding/json"
	"github.com/crawlerclub/x/types"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestUtil(t *testing.T) {
	var conf types.CrawlerConf
	bytes, err := ioutil.ReadFile("../cmd/www.newsmth.net.json")
	if err != nil {
		t.Error(err)
	}
	err = json.Unmarshal(bytes, &conf)
	if err != nil {
		t.Error(err)
	}

	ret1, err := ObjectToBytes(conf)
	if err != nil {
		t.Error(err)
	}

	var object types.CrawlerConf
	err = BytesToObject(ret1, &object)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(object, conf) {
		t.Error("object:", object)
	}
}
