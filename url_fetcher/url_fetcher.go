package main

import (
	"flag"
	"fmt"
	d "gextract/document"
)

var (
	uri = flag.String("uri", "", "the online web page")
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	flag.Parse()
	if uri == nil || len(*uri) == 0 {
		fmt.Println("url_fetcher --uri http://www.sina.com.cn")
		return
	}

	fmt.Println(d.FetchUrl2(*uri))
}
