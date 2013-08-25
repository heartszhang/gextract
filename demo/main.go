package main

import (
	"fmt"
	"gextract/feeds"
	"log"
	"os"
)

func init() {
	log.SetOutput(os.Stderr)
}
func main() {
	log.Println("hello rss.xml")
	c, items := feeds.NewRss2("rss.xml", "http://dummy.com")
	fmt.Println(c)
	feeds.InsertChannel(c)
	test_db(items)
}

func test_db(items []feeds.Entry) {
	feeds.InsertEntries(items)
	//	set := feeds.TopNEntries(0, 2)
	//	fmt.Println(set)
}