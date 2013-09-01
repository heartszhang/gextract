package main

import (
	"errors"
	"flag"
	"fmt"
	d "gextract/document"
	"gextract/feeds"
	"log"
)

var uri = flag.String("uri", "", "rss 2.0 url")

func main() {
	defer recover_panic()
	flag.Parse()
	if len(*uri) == 0 {
		fmt.Println(errors.New("entry_fetcher --uri http://feeds.feedburner.com/chinadigitaltimes/ywgz"))
		return
	}
	log.Println("fetching", *uri)
	rsfile, _, _, _ := d.DownloadFile(*uri)
	log.Println("fetch url to ", rsfile)

	f, entries, err := feeds.NewFeed(rsfile, *uri)
	log.Println(f, len(entries), err)

	feeds.NewEntryOperator().Save(entries)

}

func recover_panic() {
	if err := recover(); err != nil {
		log.Fatalln(err)
	}
}
