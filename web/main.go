package main

import (
	"fmt"
	"github.com/GeertJohan/go.rice"
	"io/ioutil"
	"net/http"
)

func addHandler(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(b))
	err = ioutil.WriteFile("./output.json", b, 0666)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "false")
	} else {
		fmt.Fprintf(w, "true")
	}
}

func main() {
	http.Handle("/", http.FileServer(rice.MustFindBox("http-files").HTTPBox()))
	http.Handle("/crawlerconf", http.HandlerFunc(addHandler))
	err := http.ListenAndServe(":8308", nil)
	if err != nil {
		fmt.Println(err)
	}
}
