package main

import (
	"flag"
	d "gextract/document"
	"gextract/feeds"
	"log"
	"os"
)

func main() {
	defer recover_panic()
	flag.Parse()
	entries := feeds.EntriesContentUnready()
	log.Println(len(entries))

	cocurrents := 12
	echans := make([]chan *feeds.Entry, cocurrents)
	done := make(chan int)

	for i := 0; i < cocurrents; i++ {
		echan := make(chan *feeds.Entry, 100)
		echans[i] = echan
		go touch_entry(echan, done)
	}

	for i, e := range entries {
		x := e
		echans[i%cocurrents] <- &x
	}

	for i := 0; i < cocurrents; i++ {
		echans[i] <- nil
	}

	for i := 0; i < cocurrents; i++ {
		<-done
	}
}

func init() {
	log.SetOutput(os.Stderr)
}

func recover_panic() {
	if err := recover(); err != nil {
		log.Fatalln(err)
	}
}

func touch_entry(echan <-chan *feeds.Entry, done chan<- int) {
	for e := <-echan; e != nil; e = <-echan {
		//		tf, _ := d.FetchUrl3(e.Link)
		tf := d.ExtractHtml(e.Link)
		feeds.EntryUpdateContent(tf, e.Link)
	}
	done <- 0
}
