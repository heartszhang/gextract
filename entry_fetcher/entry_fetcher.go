package main

import (
	"errors"
	"flag"
	d "gextract/document"
	"gextract/feeds"
	"log"
	"os"
)

var uri = flag.String("uri", "", "rss 2.0 url")

func main() {
	defer recover_panic()
	flag.Parse()
	if len(*uri) == 0 {
		panic(errors.New("usage: entry_fetcher --uri http://xome.rss2.xml"))
	}
	rsfile, _ := d.FetchUrl2(*uri)
	ch, entries := feeds.NewRss2(rsfile, *uri)
	log.Println(len(entries))
	feeds.SaveEntries(ch, entries)
	//	feeds.InsertEntries(entries)
	//	feeds.InsertChannel(ch)
}

func init() {
	log.SetOutput(os.Stderr)
}

func recover_panic() {
	if err := recover(); err != nil {
		log.Fatalln(err)
	}
}
