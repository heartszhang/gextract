package main

import (
	"fmt"
	"gextract/feeds"
	"log"
	"os"
)
var (
	rssfile = "rss.xml"
	base_url = "http://dummy.com"
)

func init() {
	log.SetOutput(os.Stderr)
}
func main() {
	c, items := feeds.NewRss2(rssfile, base_url)
	fmt.Println(c)
	feeds.InsertChannel(c)
	test_db(items)
}

func test_db(items []feeds.Entry) {
	feeds.InsertEntries(items)
}
