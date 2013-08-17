package main

import (
	"flag"
	"fmt"
	d "gextract/document"
)

var (
	url = flag.String("u", "", "the online web page")
)

func main() {
	flag.Parse()
	if url == nil || len(*url) == 0 {
		fmt.Println("like fetcher -u http://www.sina.com.cn")
		return
	}

	fmt.Println(d.FetchUrl(*url, "2.html"))
}
