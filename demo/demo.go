package main

import (
	"fmt"
	"github.com/heartszhang/gextract/feeds"
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
	c, items, _ := feeds.NewFeed(rssfile, base_url)
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
