package main

import (
	"errors"
	"flag"
	"gextract/feeds"
	d "github.com/heartszhang/gextract/document"
	"log"
	"os"
)

func init() {
	log.SetOutput(os.Stderr)
}

var uri = flag.String("uri", "", "rss 2.0 url")

func main() {
	defer catch_panic()
	flag.Parse()
	if len(*uri) == 0 {
		panic(errors.New("usage: subscribe --uri http://xome.rss2.xml"))
	}
	rsfile, _, _, _ := d.DefaultCurl().Download(*uri) //document.FetchUrl2(*uri)
	log.Println(rsfile)
	ch, _, _ := feeds.NewFeed(rsfile, *uri)
	log.Println(ch)
	feeds.NewFeedOperator().Upsert(ch)
}

func catch_panic() {
	if e := recover(); e != nil {
		log.Fatalln(e)
	}
}
