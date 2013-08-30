package main

import (
	"fmt"
	"gextract/feeds"
	"gextract/feeds/rss"
	"log"
	"os"
)

var (
	rssfile  = "rss.xml"
	base_url = "http://dummy.com"
)

func init() {
	log.SetOutput(os.Stderr)
}
func main() {
	c, items, _ := rss.NewRss2(rssfile, base_url)
	fmt.Println(c)
	for _, item := range items {
		fmt.Println(item.Title)
	}
	fo := feeds.NewFeedOperator()
	fo.Upsert(c)
	x, _ := fo.Find("http://fulltextrssfeed.com/www.infzm.com/rss/home/rss2.0.xml")
	fmt.Println(x)

	eo := feeds.NewEntryOperator()
	eo.Save(items)
	//	feeds.InsertChannel(c)
	//	test_db(items)
}

func test_db(items []feeds.Entry) {

}
