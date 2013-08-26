package main

import (
	"errors"
	"flag"
	"gextract/document"
	"gextract/feeds"
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
	rsfile := document.FetchUrl2(*uri)
	log.Println(rsfile)
	ch, _ := feeds.NewRss2(rsfile, *uri)
	log.Println(ch)
	feeds.InsertChannel(ch)
}

func catch_panic() {
	if e := recover(); e != nil {
		log.Fatalln(e)
	}
}
