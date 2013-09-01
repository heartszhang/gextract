package main

import (
	"flag"
	"fmt"
	d "github.com/heartszhang/gextract/document"
	feeds "github.com/heartszhang/gextract/feeds"
	opml "github.com/heartszhang/gextract/feeds/opml"
)

var uri = flag.String("uri", "", "opml url")

func main() {
	flag.Parse()
	if *uri == "" {
		fmt.Println("opml-parser --uri http://diggreader/opml.xml")
	}
	tf, _, _, _ := d.DefaultCurl().Download(*uri)
	ols, _ := opml.NewOpml(tf)
	feeds.NewFeedOperator().Save(ols)
	fmt.Println(ols)
}
